package xtime

import (
	"fmt"
	"strings"
	"time"
)

var defaultLayout = time.DateTime + " -0700"

// Format formats a time.Time object to a string using standard datetime format with timezone
func Format(t time.Time) string {
	return t.Format(defaultLayout)
}

// FormatLocalized formats a time.Time object using the current locale with custom template
func FormatLocalized(t time.Time, layout string) string {
	locale := GetCurrentLocale()
	return locale.FormatTemplate(t, layout)
}

// FormatMonth returns the localized month name
func FormatMonth(month time.Month, short bool) string {
	locale := GetCurrentLocale()
	return locale.FormatMonth(month, short)
}

// FormatWeekday returns the localized weekday name
func FormatWeekday(weekday time.Weekday, short bool) string {
	locale := GetCurrentLocale()
	return locale.FormatWeekday(weekday, short)
}

// FormatDuration returns a localized duration string
func FormatDuration(d time.Duration) string {
	locale := GetCurrentLocale()
	return locale.FormatDuration(d)
}

// FormatRelative returns a localized relative time string (e.g., "2 hours ago", "3 天前")
func FormatRelative(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)
	// For past times, diff is positive, but we want to show "ago"
	// For future times, diff is negative, but we want to show "from now"
	return FormatDuration(diff)
}

// FormatDateTime formats time with localized month and weekday names
func FormatDateTime(t time.Time) string {
	ft := FormatTypeDateTime
	locale := GetCurrentLocale()

	if format, ok := locale.getFormat(ft); ok {
		return FormatWithLanguage(t, locale.Language, format)
	}

	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

// FormatDate formats date with localized month and weekday names
func FormatDate(t time.Time) string {
	ft := FormatTypeDate
	locale := GetCurrentLocale()

	if format, ok := locale.getFormat(ft); ok {
		return FormatWithLanguage(t, locale.Language, format)
	}

	return fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

// FormatTime formats time using locale template
func FormatTime(t time.Time) string {
	ft := FormatTypeTime
	locale := GetCurrentLocale()

	if format, ok := locale.getFormat(ft); ok {
		return FormatWithLanguage(t, locale.Language, format)
	}

	return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
}

// FormatWithLanguage formats a time.Time object using the specified language and format type
// formatType can be: "date", "datetime", "time", or a custom template string
func FormatWithLanguage(t time.Time, language Language, format string) string {
	locale := GetCurrentLocale()

	if loc, ok := locales[language]; ok {
		locale = loc
	}

	return locale.FormatTemplate(t, format)
}

// FormatMonth returns the localized month name
func (l *Locale) FormatMonth(month time.Month, short bool) string {
	index := int(month) - 1
	if short {
		if index >= 0 && index < len(l.MonthsShort) {
			return l.MonthsShort[index]
		}
	} else {
		if index >= 0 && index < len(l.Months) {
			return l.Months[index]
		}
	}

	return month.String()
}

// FormatWeekday returns the localized weekday name
func (l *Locale) FormatWeekday(weekday time.Weekday, short bool) string {
	index := int(weekday)
	if short {
		if index >= 0 && index < len(l.WeeksShort) {
			return l.WeeksShort[index]
		}
	} else {
		if index >= 0 && index < len(l.Weeks) {
			return l.Weeks[index]
		}
	}

	return weekday.String()
}

// FormatTemplate formats a time using locale template
func (l *Locale) FormatTemplate(t time.Time, template string) string {
	if template == "" {
		return Format(t)
	}

	result := template

	// Replace placeholders with actual values
	result = strings.ReplaceAll(result, "{%y}", fmt.Sprintf("%d", t.Year()))
	result = strings.ReplaceAll(result, "{%M}", l.FormatMonth(t.Month(), false))
	result = strings.ReplaceAll(result, "{%d}", fmt.Sprintf("%d", t.Day()))
	result = strings.ReplaceAll(result, "{%w}", l.FormatWeekday(t.Weekday(), false))
	result = strings.ReplaceAll(result, "{%h}", fmt.Sprintf("%02d", t.Hour()))
	result = strings.ReplaceAll(result, "{%m}", fmt.Sprintf("%02d", t.Minute()))
	result = strings.ReplaceAll(result, "{%s}", fmt.Sprintf("%02d", t.Second()))

	return result
}

// FormatDuration returns a localized duration string
func (l *Locale) FormatDuration(d time.Duration) string {
	if d == 0 {
		return l.Now
	}

	abs := d
	if abs < 0 {
		abs = -abs
	}

	var result string

	switch {
	case abs >= 365*24*time.Hour:
		years := int(abs / (365 * 24 * time.Hour))
		result = l.formatPlural(l.Year, years)
	case abs >= 30*24*time.Hour:
		months := int(abs / (30 * 24 * time.Hour))
		result = l.formatPlural(l.Month, months)
	case abs >= 7*24*time.Hour:
		weeks := int(abs / (7 * 24 * time.Hour))
		result = l.formatPlural(l.Week, weeks)
	case abs >= 24*time.Hour:
		days := int(abs / (24 * time.Hour))
		result = l.formatPlural(l.Day, days)
	case abs >= time.Hour:
		hours := int(abs / time.Hour)
		result = l.formatPlural(l.Hour, hours)
	case abs >= time.Minute:
		minutes := int(abs / time.Minute)
		result = l.formatPlural(l.Minute, minutes)
	default:
		seconds := int(abs / time.Second)
		result = l.formatPlural(l.Second, seconds)
	}

	if d < 0 {
		return fmt.Sprintf(l.Ago, result)
	}

	return fmt.Sprintf(l.FromNow, result)
}
