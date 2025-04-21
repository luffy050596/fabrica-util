// Package camelcase provides functions to convert strings to camel case and underscore case.
package camelcase

import (
	"strings"
	"unicode"
)

// define common abbreviations (sorted by length to prevent short prefix replacement)
var commonInitialisms = []string{
	"ASCII", "MySQL",
	"XSRF", "XSS", "YAML", "UUID", "SMTP", "HTML", "HTTP", "JSON", "UTF8",
	"QPS", "CPU", "UID", "URI", "URL", "XML", "ACL", "API", "CSS", "DNS", "EOF",
	"LHS", "RAM", "RHS", "RPC", "SLA", "SSH", "TCP", "TLS", "TTL", "UDP",
	"UI", "ID", "VM", "IP",
}

var (
	camelCommonAbbrReplacer *strings.Replacer
	abbrCommonReplacer      *strings.Replacer
)

func init() {
	buildReplacers()
}

func buildReplacers() {
	camelCommonPairs := make([]string, 0, len(commonInitialisms)*2)
	abbrCommonPairs := make([]string, 0, len(commonInitialisms)*2)

	for _, abbr := range commonInitialisms {
		lower := strings.ToLower(abbr)
		camel := []rune(lower)
		camel[0] = unicode.ToUpper(camel[0])
		camelCommonPairs = append(camelCommonPairs, string(camel), abbr)
		abbrCommonPairs = append(abbrCommonPairs, abbr, string(camel))
	}

	camelCommonAbbrReplacer = strings.NewReplacer(camelCommonPairs...)
	abbrCommonReplacer = strings.NewReplacer(abbrCommonPairs...)
}

// ToUpperCamel converts a string to upper camel case.
func ToUpperCamel(s string) string {
	if s == "" {
		return ""
	}

	s = toUpperCamel(s)
	s = camelCommonAbbrReplacer.Replace(s)

	return s
}

// ToLowerCamel converts a string to lower camel case.
func ToLowerCamel(s string) string {
	if s == "" {
		return ""
	}

	s = toUpperCamel(s)
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	s = string(r)
	s = camelCommonAbbrReplacer.Replace(s)

	return s
}

func toUpperCamel(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")

	if len(parts) == 1 {
		r := []rune(s)
		r[0] = unicode.ToUpper(r[0])

		return string(r)
	}

	var builder strings.Builder

	for _, p := range parts {
		if p == "" {
			continue
		}

		r := []rune(strings.ToLower(p))
		r[0] = unicode.ToUpper(r[0])
		builder.WriteString(string(r))
	}

	return builder.String()
}

// ToUnderScore converts a string to underscore case.
func ToUnderScore(s string) string {
	if s == "" {
		return ""
	}

	s = abbrCommonReplacer.Replace(s)

	var builder strings.Builder

	runes := []rune(s)
	length := len(runes)

	for i := range length {
		// is upper or digit
		if unicode.IsUpper(runes[i]) || unicode.IsDigit(runes[i]) {
			// is not first character and previous character is not upper
			if i > 0 &&
				!unicode.IsUpper(runes[i-1]) &&
				!unicode.IsDigit(runes[i-1]) {
				builder.WriteByte('_')
			}

			builder.WriteRune(unicode.ToLower(runes[i]))
		} else {
			builder.WriteRune(runes[i])
		}
	}

	return builder.String()
}
