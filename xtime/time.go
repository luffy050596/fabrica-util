// Package xtime provides extended time utilities for formatting and calculating time periods
package xtime

import (
	"log/slog"
	"time"

	"github.com/go-pantheon/fabrica-util/errors"
)

var (
	location *time.Location
)

// Config represents the configuration for time package
type Config struct {
	Language Language
	Timezone string
}

// Init initializes the time package with the specified configuration
func Init(cfg Config) error {
	if cfg.Timezone == "" {
		cfg.Timezone = "UTC"
	}

	if cfg.Language == "" {
		cfg.Language = "en"
	}

	// Load timezone
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return errors.Wrapf(err, "failed to load timezone %s", cfg.Timezone)
	}

	location = loc

	code, ok := parseLanguageCode(string(cfg.Language))
	if !ok {
		return errors.Errorf("invalid language code: %s", cfg.Language)
	}

	// Load locale
	if err := SetLocale(code); err != nil {
		slog.Error("failed to load locale", "language", cfg.Language, "error", err)
		// If locale loading fails, continue with default English locale
		// but don't return error as this shouldn't break the time functionality
		initDefaultLocale()
	}

	return nil
}

// InitSimple initializes the time package with language only (for backward compatibility)
func InitSimple(language string) error {
	return Init(Config{Language: Language(language)})
}

// Time converts a timestamp to a time.Time object
// Returns zero time if timestamp is 0
func Time(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}

	return time.Unix(timestamp, 0).UTC()
}

// NextDailyTime calculates the next daily time after the given time with the specified delay
func NextDailyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfDay(t.Add(-delay)).AddDate(0, 0, 1).Add(delay)
}

// NextWeeklyTime calculates the next weekly time after the given time with the specified delay
func NextWeeklyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfWeek(t.Add(-delay)).AddDate(0, 0, 7).Add(delay)
}

// NextMonthlyTime calculates the next monthly time after the given time with the specified delay
func NextMonthlyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfMonth(t.Add(-delay)).AddDate(0, 1, 0).Add(delay)
}

// StartOfDay returns the start of the day for the given time
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// StartOfWeek returns the start of the week for the given time (Monday as first day)
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}

	daysToSubtract := weekday - 1

	return StartOfDay(t.AddDate(0, 0, -daysToSubtract))
}

// StartOfMonth returns the start of the month for the given time
func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

// InTimezone converts time to specified timezone
func InTimezone(t time.Time, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "failed to load timezone %s", tz)
	}

	return t.In(loc), nil
}

// GetLocation returns the current location, UTC if not initialized
func GetLocation() *time.Location {
	if location != nil {
		return location
	}

	return time.UTC
}
