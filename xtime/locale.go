// Package xtime locale support for multi-language formatting
package xtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-pantheon/fabrica-util/errors"
)

func init() {
	initDefaultLocale()
}

// Language represents a language code
type Language string

const (
	// LanguageEn is the English language code
	LanguageEn Language = "en"
	// LanguageZhCN is the Chinese Simplified language code
	LanguageZhCN Language = "zh-CN"
	// LanguageZhTW is the Chinese Traditional language code
	LanguageZhTW Language = "zh-TW"
	// LanguageJp is the Japanese language code
	LanguageJp Language = "jp"
	// LanguageKr is the Korean language code
	LanguageKr Language = "kr"
)

// validLanguageCodes defines the allowed language codes to prevent path traversal
var validLanguageCodes = map[string]Language{
	string(LanguageEn):   LanguageEn,
	string(LanguageZhCN): LanguageZhCN,
	string(LanguageZhTW): LanguageZhTW,
	string(LanguageJp):   LanguageJp,
	string(LanguageKr):   LanguageKr,
}

// FormatType represents a format type
type FormatType string

// FormatTypeDate is the date format type
const (
	// FormatTypeDate is the date format type
	FormatTypeDate FormatType = "date"
	// FormatTypeDateTime is the date and time format type
	FormatTypeDateTime FormatType = "datetime"
	// FormatTypeTime is the time format type
	FormatTypeTime FormatType = "time"
)

// Locale represents a language locale configuration
type Locale struct {
	Language       Language
	Months         []string `json:"-"`
	MonthsShort    []string `json:"-"`
	Weeks          []string `json:"-"`
	WeeksShort     []string `json:"-"`
	Constellations []string `json:"-"`

	Format map[FormatType]string `json:"format"`

	// Duration formats
	Year    string `json:"year"`
	Month   string `json:"month"`
	Week    string `json:"week"`
	Day     string `json:"day"`
	Hour    string `json:"hour"`
	Minute  string `json:"minute"`
	Second  string `json:"second"`
	Now     string `json:"now"`
	Ago     string `json:"ago"`
	FromNow string `json:"from_now"`
	Before  string `json:"before"`
	After   string `json:"after"`
}

// localeData is used for JSON unmarshaling
type localeData struct {
	Months         string                `json:"months"`
	MonthsShort    string                `json:"months_short"`
	Weeks          string                `json:"weeks"`
	WeeksShort     string                `json:"weeks_short"`
	Constellations string                `json:"constellations"`
	Format         map[FormatType]string `json:"format"`
	Year           string                `json:"year"`
	Month          string                `json:"month"`
	Week           string                `json:"week"`
	Day            string                `json:"day"`
	Hour           string                `json:"hour"`
	Minute         string                `json:"minute"`
	Second         string                `json:"second"`
	Now            string                `json:"now"`
	Ago            string                `json:"ago"`
	FromNow        string                `json:"from_now"`
	Before         string                `json:"before"`
	After          string                `json:"after"`
}

var (
	currentLocale *Locale
	locales       = make(map[Language]*Locale)
)

// languageCodePattern matches valid language codes
var languageCodePattern = regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)

// parseLanguageCode parses the language code and returns the Language enum
func parseLanguageCode(language string) (Language, bool) {
	if !languageCodePattern.MatchString(language) {
		return "", false
	}

	code, ok := validLanguageCodes[language]

	return code, ok
}

// sanitizeAndBuildPath safely constructs the path to the language file
func sanitizeAndBuildPath(language string) string {
	// Double check the language is valid
	_, ok := parseLanguageCode(language)
	if !ok {
		return ""
	}

	// Construct safe path - no user input is directly used in path construction
	filename := language + ".json"

	// Try local lang directory first
	langFile := filepath.Join("lang", filename)
	if _, err := os.Stat(langFile); err == nil {
		return langFile
	}

	// Try relative to the package directory
	dir, _ := os.Getwd()

	langFile = filepath.Join(dir, "deps", "fabrica-util", "xtime", "lang", filename)
	if _, err := os.Stat(langFile); err == nil {
		return langFile
	}

	return ""
}

// initDefaultLocale initializes the default English locale
func initDefaultLocale() {
	defaultLocale := &Locale{
		Language:       LanguageEn,
		Months:         []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		MonthsShort:    []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		Weeks:          []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		WeeksShort:     []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		Constellations: []string{"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo", "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"},
		Format: map[FormatType]string{
			FormatTypeDate:     "{%w}, {%M} {%d}, {%y}",
			FormatTypeDateTime: "{%w}, {%M} {%d}, {%y} {%h}:{%m}:{%s}",
			FormatTypeTime:     "{%h}:{%m}:{%s}",
		},
		Year:    "1 year|%d years",
		Month:   "1 month|%d months",
		Week:    "1 week|%d weeks",
		Day:     "1 day|%d days",
		Hour:    "1 hour|%d hours",
		Minute:  "1 minute|%d minutes",
		Second:  "1 second|%d seconds",
		Now:     "just now",
		Ago:     "%s ago",
		FromNow: "%s from now",
		Before:  "%s before",
		After:   "%s after",
	}

	locales["en"] = defaultLocale
	currentLocale = defaultLocale
}

// LoadLocale loads a locale from JSON file
func LoadLocale(language Language) error {
	// Check if already loaded
	if locale, exists := locales[language]; exists {
		currentLocale = locale
		return nil
	}

	// Validate language to prevent path traversal attacks
	_, ok := parseLanguageCode(string(language))
	if !ok {
		return fmt.Errorf("invalid language code: %s", language)
	}

	// Use embedded files or safe path construction
	langFile := sanitizeAndBuildPath(string(language))
	if langFile == "" {
		return fmt.Errorf("locale file not found for language: %s", language)
	}

	// Read and parse the JSON file
	data, err := os.ReadFile(filepath.Clean(langFile))
	if err != nil {
		return errors.Wrapf(err, "failed to read locale file: %s", langFile)
	}

	var localeData localeData
	if err := json.Unmarshal(data, &localeData); err != nil {
		return errors.Wrapf(err, "failed to parse locale file: %s", langFile)
	}

	// Convert to Locale struct
	locale := &Locale{
		Language:       language,
		Months:         strings.Split(localeData.Months, "|"),
		MonthsShort:    strings.Split(localeData.MonthsShort, "|"),
		Weeks:          strings.Split(localeData.Weeks, "|"),
		WeeksShort:     strings.Split(localeData.WeeksShort, "|"),
		Constellations: strings.Split(localeData.Constellations, "|"),
		Format: map[FormatType]string{
			FormatTypeDate:     localeData.Format[FormatTypeDate],
			FormatTypeDateTime: localeData.Format[FormatTypeDateTime],
			FormatTypeTime:     localeData.Format[FormatTypeTime],
		},
		Year:    localeData.Year,
		Month:   localeData.Month,
		Week:    localeData.Week,
		Day:     localeData.Day,
		Hour:    localeData.Hour,
		Minute:  localeData.Minute,
		Second:  localeData.Second,
		Now:     localeData.Now,
		Ago:     localeData.Ago,
		FromNow: localeData.FromNow,
		Before:  localeData.Before,
		After:   localeData.After,
	}

	locales[language] = locale
	currentLocale = locale

	return nil
}

var localOnce sync.Once

// GetCurrentLocale returns the current locale
func GetCurrentLocale() *Locale {
	if currentLocale == nil {
		localOnce.Do(initDefaultLocale)
	}

	return currentLocale
}

// SetLocale sets the current locale
func SetLocale(language Language) error {
	if locale, exists := locales[language]; exists {
		currentLocale = locale
		return nil
	}

	return LoadLocale(language)
}

// GetAvailableLanguages returns a list of available languages
func GetAvailableLanguages() []Language {
	languages := make([]Language, 0, len(locales))
	for lang := range locales {
		languages = append(languages, lang)
	}

	return languages
}

// formatPlural formats plural strings based on count
func (l *Locale) formatPlural(format string, count int) string {
	parts := strings.Split(format, "|")
	if len(parts) == 1 {
		// Simple format like "%d 年"
		return fmt.Sprintf(format, count)
	}

	// English-style plural: "1 year|%d years"
	if count == 1 && len(parts) >= 1 {
		return parts[0]
	}

	if len(parts) >= 2 {
		return fmt.Sprintf(parts[1], count)
	}

	return fmt.Sprintf(format, count)
}

func (l *Locale) getFormat(formatType FormatType) (string, bool) {
	if format, ok := l.Format[formatType]; ok {
		return format, true
	}

	return "", false
}
