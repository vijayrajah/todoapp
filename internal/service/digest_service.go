package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"todoapp/internal/model"
	"todoapp/internal/store"
)

const (
	digestCheckEvery = time.Minute
	digestRetryAfter = 15 * time.Minute
	digestMaxListed  = 20
)

// Digest is the summary of overdue and due-today tasks.
type Digest struct {
	Date     string       `json:"date"`
	Overdue  []model.Task `json:"overdue"`
	DueToday []model.Task `json:"due_today"`
	Show     bool         `json:"show"` // the in-app digest should pop up now
	Text     string       `json:"text"`
}

// DigestService builds the daily digest and delivers it to an optional webhook.
type DigestService struct {
	store  *store.Store
	log    zerolog.Logger
	client *http.Client
	now    func() time.Time

	mu          sync.Mutex
	lastAttempt time.Time
}

// NewDigestService returns a DigestService.
func NewDigestService(st *store.Store, log zerolog.Logger) *DigestService {
	return &DigestService{store: st, log: log, client: &http.Client{Timeout: 10 * time.Second}, now: time.Now}
}

// Build returns the digest for today without deciding whether to show it.
func (s *DigestService) Build(ctx context.Context) (Digest, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return Digest{}, err
	}
	today := dayOf(s.now(), profile)
	overdue, err := s.store.ListTasks(ctx, model.TaskFilter{
		Status: model.StatusTodo, DueTo: dateString(today.AddDate(0, 0, -1)), Order: "due"})
	if err != nil {
		return Digest{}, err
	}
	dueToday, err := s.store.ListTasks(ctx, model.TaskFilter{
		Status: model.StatusTodo, DueFrom: dateString(today), DueTo: dateString(today), Order: "due"})
	if err != nil {
		return Digest{}, err
	}
	d := Digest{Date: dateString(today), Overdue: overdue, DueToday: dueToday}
	d.Text = digestText(d)
	return d, nil
}

// Current returns today's digest, with Show set when the in-app digest is enabled, its time has
// passed, it has not been shown today and there is something to report.
func (s *DigestService) Current(ctx context.Context) (Digest, error) {
	d, err := s.Build(ctx)
	if err != nil {
		return Digest{}, err
	}
	prefs, err := s.store.GetPreferences(ctx)
	if err != nil {
		return Digest{}, err
	}
	shown, err := s.store.DigestShownOn(ctx)
	if err != nil {
		return Digest{}, err
	}
	d.Show = prefs.DigestEnabled && s.timePassed(ctx, prefs) && shown != d.Date && len(d.Overdue)+len(d.DueToday) > 0
	return d, nil
}

// MarkShown records that today's in-app digest was shown.
func (s *DigestService) MarkShown(ctx context.Context) error {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return err
	}
	return s.store.MarkDigestShown(ctx, TodayString(s.now(), profile))
}

func (s *DigestService) timePassed(ctx context.Context, prefs model.Preferences) bool {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return false
	}
	return s.now().In(profile.Location()).Format("15:04") >= prefs.DigestTime
}

// SendNow posts the digest to the configured webhook, regardless of schedule.
func (s *DigestService) SendNow(ctx context.Context) error {
	prefs, err := s.store.GetPreferences(ctx)
	if err != nil {
		return err
	}
	if prefs.DigestWebhook == "" {
		return model.ValidationError("set a webhook URL first")
	}
	d, err := s.Build(ctx)
	if err != nil {
		return err
	}
	return s.post(ctx, prefs.DigestWebhook, d.Text)
}

func (s *DigestService) post(ctx context.Context, url, text string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(text))
	if err != nil {
		return model.ValidationError("invalid webhook URL")
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Title", "todoapp daily digest") // understood by ntfy
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send digest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("send digest: webhook returned %s", resp.Status)
	}
	return nil
}

// Start runs the webhook scheduler until ctx is cancelled.
func (s *DigestService) Start(ctx context.Context) {
	go func() {
		t := time.NewTicker(digestCheckEvery)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.tick(ctx)
			}
		}
	}()
}

func (s *DigestService) tick(ctx context.Context) {
	prefs, err := s.store.GetPreferences(ctx)
	if err != nil || !prefs.DigestEnabled || prefs.DigestWebhook == "" || !s.timePassed(ctx, prefs) {
		return
	}
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return
	}
	today := TodayString(s.now(), profile)
	if prefs.DigestLastSent == today {
		return
	}
	s.mu.Lock()
	recent := !s.lastAttempt.IsZero() && s.now().Sub(s.lastAttempt) < digestRetryAfter
	if !recent {
		s.lastAttempt = s.now()
	}
	s.mu.Unlock()
	if recent {
		return
	}

	d, err := s.Build(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("digest: build")
		return
	}
	if len(d.Overdue)+len(d.DueToday) > 0 {
		if err := s.post(ctx, prefs.DigestWebhook, d.Text); err != nil {
			s.log.Error().Err(err).Msg("digest: webhook failed")
			return
		}
		s.log.Info().Msg("daily digest sent")
	}
	if err := s.store.MarkDigestSent(ctx, today); err != nil {
		s.log.Error().Err(err).Msg("digest: record sent")
	}
}

func digestText(d Digest) string {
	if len(d.Overdue)+len(d.DueToday) == 0 {
		return "Nothing overdue or due today."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d overdue, %d due today (%s)\n", len(d.Overdue), len(d.DueToday), d.Date)
	section := func(title string, tasks []model.Task, withDue bool) {
		if len(tasks) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n%s:\n", title)
		for i, t := range tasks {
			if i == digestMaxListed {
				fmt.Fprintf(&b, "...and %d more\n", len(tasks)-digestMaxListed)
				break
			}
			if withDue {
				fmt.Fprintf(&b, "- %s (due %s)\n", t.Title, t.DueDate)
			} else {
				fmt.Fprintf(&b, "- %s\n", t.Title)
			}
		}
	}
	section("Overdue", d.Overdue, true)
	section("Due today", d.DueToday, false)
	return b.String()
}
