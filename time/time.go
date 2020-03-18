package time

import (
	"time"
)

func Now() time.Time {
	return time.Now()
}

func NowUnix() int64 {
	return time.Now().Unix()
}

func Time(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}

func NextDailyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfDay(t.Add(-delay)).AddDate(0, 0, 1).Add(delay)
}

func NextWeeklyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfWeek(t.Add(-delay)).AddDate(0, 0, 7).Add(delay)
}

func NextMonthlyTime(t time.Time, delay time.Duration) time.Time {
	return StartOfMonth(t.Add(-delay)).AddDate(0, 1, 0).Add(delay)
}

func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func StartOfWeek(t time.Time) time.Time {
	year, month, day := t.Date()
	daysSinceMonday := int(t.Weekday())
	if daysSinceMonday == 0 {
		daysSinceMonday = 7
	}
	daysSinceMonday--
	return time.Date(year, month, day-daysSinceMonday, 0, 0, 0, 0, t.Location())
}

func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}
