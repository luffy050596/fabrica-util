package xtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleInit(t *testing.T) {
	t.Parallel()

	// Test that default locale is English
	locale := GetCurrentLocale()
	assert.NotNil(t, locale)
	assert.Equal(t, LanguageEn, locale.Language)
	assert.Equal(t, "January", locale.Months[0])
	assert.Equal(t, "Jan", locale.MonthsShort[0])
}

//nolint:paralleltest // modifies global locale state
func TestSetLocale(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		wantErr  bool
	}{
		{
			name:     "set to default English",
			language: LanguageEn,
			wantErr:  false,
		},
		{
			name:     "invalid language",
			language: Language("invalid"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetLocale(tt.language)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatMonth(t *testing.T) {
	tests := []struct {
		month    time.Month
		short    bool
		expected string
	}{
		{time.January, false, "January"},
		{time.January, true, "Jan"},
		{time.December, false, "December"},
		{time.December, true, "Dec"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// Set English locale for each test
			err := SetLocale("en")
			require.NoError(t, err)

			result := FormatMonth(tt.month, tt.short)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatWeekday(t *testing.T) {
	tests := []struct {
		weekday  time.Weekday
		short    bool
		expected string
	}{
		{time.Sunday, false, "Sunday"},
		{time.Sunday, true, "Sun"},
		{time.Monday, false, "Monday"},
		{time.Monday, true, "Mon"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// Set English locale for each test
			err := SetLocale("en")
			require.NoError(t, err)

			result := FormatWeekday(tt.weekday, tt.short)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "zero duration",
			duration: 0,
			expected: "just now",
		},
		{
			name:     "1 second",
			duration: 1 * time.Second,
			expected: "1 second from now",
		},
		{
			name:     "30 seconds",
			duration: 30 * time.Second,
			expected: "30 seconds from now",
		},
		{
			name:     "1 minute",
			duration: 1 * time.Minute,
			expected: "1 minute from now",
		},
		{
			name:     "2 minutes",
			duration: 2 * time.Minute,
			expected: "2 minutes from now",
		},
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
			expected: "1 hour from now",
		},
		{
			name:     "2 hours",
			duration: 2 * time.Hour,
			expected: "2 hours from now",
		},
		{
			name:     "1 day",
			duration: 24 * time.Hour,
			expected: "1 day from now",
		},
		{
			name:     "2 days",
			duration: 48 * time.Hour,
			expected: "2 days from now",
		},
		{
			name:     "past duration",
			duration: -2 * time.Hour,
			expected: "2 hours ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set English locale for each test
			err := SetLocale("en")
			require.NoError(t, err)

			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatDateTime(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		language Language
		expected string
	}{
		{
			language: LanguageEn,
			expected: "Monday, December 25, 2023 15:30:45",
		},
		{
			language: LanguageZhCN,
			expected: "2023年十二月25日 星期一 15:30:45",
		},
		{
			language: LanguageZhTW,
			expected: "2023年十二月25日 星期一 15:30:45",
		},
		{
			language: LanguageJp,
			expected: "2023年十二月25日 月曜日 15:30:45",
		},
		{
			language: LanguageKr,
			expected: "2023년 십이월 25일 월요일 15:30:45",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.language), func(t *testing.T) {
			err := SetLocale(tt.language)
			require.NoError(t, err)

			result := FormatDateTime(testTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatDate(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		language Language
		expected string
	}{
		{
			language: LanguageEn,
			expected: "Monday, December 25, 2023",
		},
		{
			language: LanguageZhCN,
			expected: "2023年十二月25日 星期一",
		},
		{
			language: LanguageZhTW,
			expected: "2023年十二月25日 星期一",
		},
		{
			language: LanguageJp,
			expected: "2023年十二月25日 月曜日",
		},
		{
			language: LanguageKr,
			expected: "2023년 십이월 25일 월요일",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.language), func(t *testing.T) {
			err := SetLocale(tt.language)
			require.NoError(t, err)

			result := FormatDate(testTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatTime(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		language Language
		expected string
	}{
		{
			language: LanguageEn,
			expected: "15:30:45",
		},
		{
			language: LanguageZhCN,
			expected: "15:30:45",
		},
		{
			language: LanguageJp,
			expected: "15:30:45",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.language), func(t *testing.T) {
			err := SetLocale(tt.language)
			require.NoError(t, err)

			result := FormatTime(testTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // modifies global locale state
func TestFormatRelative(t *testing.T) {
	// Test with English locale
	err := SetLocale("en")
	require.NoError(t, err)

	now := time.Now()
	pastTime := now.Add(-2 * time.Hour)
	futureTime := now.Add(3 * time.Hour)

	// Test past time
	pastResult := FormatRelative(pastTime)
	assert.Contains(t, pastResult, "ago")
	assert.Contains(t, pastResult, "hours")

	// Test future time
	futureResult := FormatRelative(futureTime)
	assert.Contains(t, futureResult, "from now")
	assert.Contains(t, futureResult, "hours")
}

//nolint:paralleltest // modifies global locale state
func TestInitWithLanguage(t *testing.T) {
	err := Init(Config{
		Language: LanguageEn,
		Timezone: "UTC",
	})
	require.NoError(t, err)

	locale := GetCurrentLocale()
	assert.Equal(t, LanguageEn, locale.Language)
}

func TestGetAvailableLanguages(t *testing.T) {
	t.Parallel()

	languages := GetAvailableLanguages()
	assert.Contains(t, languages, LanguageEn)
	assert.True(t, len(languages) >= 1)
}

//nolint:paralleltest // modifies global locale state
func TestFormatWithLanguage(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name       string
		language   Language
		formatType FormatType
		format     string
		expected   string
		wantErr    bool
	}{
		{
			name:       "English date",
			language:   LanguageEn,
			formatType: FormatTypeDate,
			expected:   "Monday, December 25, 2023",
			wantErr:    false,
		},
		{
			name:       "Chinese datetime",
			language:   LanguageZhCN,
			formatType: FormatTypeDateTime,
			expected:   "2023年十二月25日 星期一 15:30:45",
			wantErr:    false,
		},
		{
			name:       "Japanese time",
			language:   LanguageJp,
			formatType: FormatTypeTime,
			expected:   "15:30:45",
			wantErr:    false,
		},
		{
			name:     "Custom template",
			language: LanguageZhCN,
			format:   "今天是{%y}年{%M}{%d}日",
			expected: "今天是2023年十二月25日",
			wantErr:  false,
		},
		{
			name:     "Go layout format",
			language: LanguageEn,
			expected: "2023-12-25 15:30:45 +0000",
			wantErr:  false,
		},
		{
			name:     "Invalid language",
			language: Language("invalid"),
			expected: "2023-12-25 15:30:45 +0000",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetLocale(tt.language); err != nil {
				_ = SetLocale("en")
			}

			format := tt.format
			if format == "" {
				if f, ok := GetCurrentLocale().getFormat(tt.formatType); ok {
					format = f
				}
			}

			result := FormatWithLanguage(testTime, tt.language, format)
			assert.Equal(t, tt.expected, result)
		})
	}
}
