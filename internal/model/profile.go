package model

import (
	"strings"
	"time"
)

// Profile holds the single user's details.
type Profile struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Timezone    string `json:"timezone"`
	Bio         string `json:"bio"`
	AvatarColor string `json:"avatar_color"`
}

// Clean trims and validates user-editable fields.
func (p *Profile) Clean() error {
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	p.Bio = strings.TrimSpace(p.Bio)
	if len(p.Name) > 100 || len(p.Email) > 200 || len(p.Bio) > 2000 {
		return ValidationError("profile field is too long")
	}
	if p.Timezone == "" {
		p.Timezone = "UTC"
	}
	if _, err := time.LoadLocation(p.Timezone); err != nil {
		return ValidationError("unknown timezone")
	}
	if p.AvatarColor == "" {
		p.AvatarColor = "#5b6cf0"
	}
	if !ValidColor(p.AvatarColor) {
		return ValidationError("avatar color must be #rrggbb")
	}
	return nil
}

// Location returns the profile time zone, falling back to UTC.
func (p Profile) Location() *time.Location {
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
