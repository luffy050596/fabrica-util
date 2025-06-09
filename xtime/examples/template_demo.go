// Package main provides an example of using the xtime package
package main

import (
	"fmt"
	"time"

	"github.com/go-pantheon/fabrica-util/xtime"
)

/*
Output:
```bash
=== Date/Time Formatting with Locale Templates ===
Test Time: 2023-12-25T15:30:45Z

Language: en
---
FormatDateTime: Monday, December 25, 2023 15:30:45
FormatDate:     Monday, December 25, 2023
FormatTime:     15:30:45
FormatMonth:    December
FormatWeekday:  Monday

Language: zh-CN
---
FormatDateTime: 2023年十二月25日 星期一 15:30:45
FormatDate:     2023年十二月25日 星期一
FormatTime:     15:30:45
FormatMonth:    十二月
FormatWeekday:  星期一

Language: zh-TW
---
FormatDateTime: 2023年十二月25日 星期一 15:30:45
FormatDate:     2023年十二月25日 星期一
FormatTime:     15:30:45
FormatMonth:    十二月
FormatWeekday:  星期一

Language: jp
---
FormatDateTime: 2023年十二月25日 月曜日 15:30:45
FormatDate:     2023年十二月25日 月曜日
FormatTime:     15:30:45
FormatMonth:    十二月
FormatWeekday:  月曜日

Language: kr
---
FormatDateTime: 2023년 십이월 25일 월요일 15:30:45
FormatDate:     2023년 십이월 25일 월요일
FormatTime:     15:30:45
FormatMonth:    십이월
FormatWeekday:  월요일

=== Custom Template Example ===
Custom Chinese format: 今天是2023年十二月25日，星期一
Custom English format: Today is Monday, December 25, 2023

=== FormatWithLanguage Demo ===
English Date             : date
Chinese DateTime         : datetime
Japanese Time            : time
Korean Custom Template   : 2023년 십이월 25일
Go Layout Format         : 2006-01-02
```
*/
func main() {
	// Initialize with different languages
	languages := []xtime.Language{
		xtime.LanguageEn,
		xtime.LanguageZhCN,
		xtime.LanguageZhTW,
		xtime.LanguageJp,
		xtime.LanguageKr,
	}

	// Test time
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	fmt.Println("=== Date/Time Formatting with Locale Templates ===")
	fmt.Printf("Test Time: %s\n\n", testTime.Format(time.RFC3339))

	for _, lang := range languages {
		fmt.Printf("Language: %s\n", lang)
		fmt.Println("---")

		// Initialize with language
		err := xtime.Init(xtime.Config{Language: lang, Timezone: "UTC"})
		if err != nil {
			fmt.Printf("Error initializing %s: %v\n", lang, err)
			continue
		}

		// Test different formatting functions
		fmt.Printf("FormatDateTime: %s\n", xtime.FormatDateTime(testTime))
		fmt.Printf("FormatDate:     %s\n", xtime.FormatDate(testTime))
		fmt.Printf("FormatTime:     %s\n", xtime.FormatTime(testTime))
		fmt.Printf("FormatMonth:    %s\n", xtime.FormatMonth(testTime.Month(), false))
		fmt.Printf("FormatWeekday:  %s\n", xtime.FormatWeekday(testTime.Weekday(), false))

		fmt.Println()
	}

	fmt.Println("=== Custom Template Example ===")
	// Example of using FormatLocalized with custom template
	err := xtime.SetLocale("zh-CN")
	if err == nil {
		customFormat := xtime.FormatLocalized(testTime, "今天是{%y}年{%M}{%d}日，{%w}")
		fmt.Printf("Custom Chinese format: %s\n", customFormat)
	}

	err = xtime.SetLocale("en")
	if err == nil {
		customFormat := xtime.FormatLocalized(testTime, "Today is {%w}, {%M} {%d}, {%y}")
		fmt.Printf("Custom English format: %s\n", customFormat)
	}

	fmt.Println("\n=== FormatWithLanguage Demo ===")

	// Test FormatWithLanguage function
	testCases := []struct {
		language    xtime.Language
		formatType  xtime.FormatType
		description string
	}{
		{"en", "date", "English Date"},
		{"zh-CN", "datetime", "Chinese DateTime"},
		{"jp", "time", "Japanese Time"},
		{"kr", "{%y}년 {%M} {%d}일", "Korean Custom Template"},
		{"en", "2006-01-02", "Go Layout Format"},
	}

	for _, tc := range testCases {
		result := xtime.FormatWithLanguage(testTime, tc.language, string(tc.formatType))
		fmt.Printf("%-25s: %s\n", tc.description, result)
	}
}
