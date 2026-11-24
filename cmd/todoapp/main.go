package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"todoapp/internal/buildinfo"
	"todoapp/internal/config"
	"todoapp/internal/daemon"
	"todoapp/internal/db"
	"todoapp/internal/logging"
	"todoapp/internal/service"
	"todoapp/internal/store"
	"todoapp/internal/web"
	"todoapp/ui"
)

// Injected at build time: -ldflags "-X main.buildTime=... -X main.gitCommit=...".
var (
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// withoutDaemonFlag drops --daemon (and --daemon=...) so the child does not daemonize again.
func withoutDaemonFlag(args []string) []string {
	var out []string
	for _, a := range args {
		if a == "--daemon" || strings.HasPrefix(a, "--daemon=") {
			continue
		}
		out = append(out, a)
	}
	return out
}

func run(args []string, out io.Writer) error {
	info := buildinfo.New(buildTime, gitCommit)

	cfg, err := config.Parse(args, info, out)
	if errors.Is(err, config.ErrExit) {
		return nil
	}
	if err != nil {
		return err
	}

	if cfg.Daemon && !daemon.IsChild() {
		if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
			return err
		}
		pid, err := daemon.Start(withoutDaemonFlag(args), cfg.DaemonLogFile())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "todoapp started in the background (pid %d), logging to %s\n", pid, cfg.DaemonLogFile())
		return nil
	}

	// The background process already has stdout/stderr redirected to the log file.
	logFile := cfg.LogFile
	if daemon.IsChild() {
		logFile = ""
	}
	log, closeLog, err := logging.New(cfg.LogLevel, cfg.LogJSON, logFile)
	if err != nil {
		return err
	}
	defer closeLog()

	if cfg.PIDFile != "" {
		removePID, err := daemon.WritePID(cfg.PIDFile)
		if err != nil {
			return err
		}
		defer removePID()
	}

	conn, err := db.Open(cfg.DBPath())
	if err != nil {
		return err
	}
	defer conn.Close()

	st := store.New(conn)
	backups := service.NewBackupService(st, cfg.DefaultBackupDir(), log)
	digest := service.NewDigestService(st, log)
	srv, err := web.NewServer(web.Options{
		Store: st, UI: ui.FS, Log: log, Build: info, ReadOnly: cfg.ReadOnly, Backups: backups, Digest: digest,
	})
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if !cfg.ReadOnly {
		backups.Start(ctx)
		digest.Start(ctx)
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.ListenAndServe() }()
	log.Info().Str("addr", cfg.Addr).Str("db", cfg.DBPath()).Bool("read_only", cfg.ReadOnly).
		Str("commit", info.GitCommit).Msg("todoapp started")

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		log.Info().Msg("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}
