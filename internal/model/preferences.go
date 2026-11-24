package model

import (
	"net/url"
	"regexp"
	"strings"
)

// Themes the UI offers. "auto" follows the operating system.
var Themes = []string{"auto", "light", "dark", "midnight", "nord", "solarized", "sepia"}

var timeOfDayRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// Preferences holds appearance and daily digest settings (single row).
type Preferences struct {
	Theme          string `json:"theme"`
	Accent         string `json:"accent"` // "" = theme default, else #rrggbb
	DigestEnabled  bool   `json:"digest_enabled"`
	DigestTime     string `json:"digest_time"` // HH:MM in the profile time zone
	DigestWebhook  string `json:"digest_webhook"`
	DigestLastSent string `json:"digest_last_sent"` // read-only (date)
}

// CleanAppearance validates theme and accent.
func (p *Preferences) CleanAppearance() error {
	ok := false
	for _, t := range Themes {
		if p.Theme == t {
			ok = true
		}
	}
	if !ok {
		return ValidationError("unknown theme")
	}
	if p.Accent != "" && !ValidColor(p.Accent) {
		return ValidationError("accent colour must be #rrggbb")
	}
	p.Accent = strings.ToLower(p.Accent)
	return nil
}

// CleanDigest validates digest settings.
func (p *Preferences) CleanDigest() error {
	if !timeOfDayRe.MatchString(p.DigestTime) {
		return ValidationError("digest time must be HH:MM")
	}
	p.DigestWebhook = strings.TrimSpace(p.DigestWebhook)
	if p.DigestWebhook != "" {
		u, err := url.Parse(p.DigestWebhook)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(p.DigestWebhook) > 500 {
			return ValidationError("webhook must be an http(s) URL")
		}
	}
	return nil
}
