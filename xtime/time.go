// Package xtime provides extended time utilities for formatting and calculating time periods
package xtime

import (
	"time"

	"github.com/dromara/carbon/v2"
)

var (
	c *carbon.Carbon
)

// Init initializes the time package with the specified language for formatting
func Init(language string) {
	c = carbon.NewCarbon().SetTimezone(carbon.DefaultTimezone)
	c.SetLanguage(carbon.NewLanguage().SetLocale(language))
}

// Time converts a timestamp to a time.Time object
// Returns zero time if timestamp is 0
func Time(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}

	return carbon.CreateFromTimestamp(timestamp, c.Timezone()).StdTime()
}

// Format formats a time.Time object to a string using standard datetime format with timezone
func Format(t time.Time) string {
	return carbon.CreateFromStdTime(t, c.Timezone()).Format(time.DateTime + " -0700")
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
	return t.Truncate(24 * time.Hour)
}

// StartOfWeek returns the start of the week for the given time
func StartOfWeek(t time.Time) time.Time {
	return t.Truncate(7 * 24 * time.Hour)
}

// StartOfMonth returns the start of the month for the given time
func StartOfMonth(t time.Time) time.Time {
	return carbon.CreateFromStdTime(t).StartOfMonth().StdTime()
}
