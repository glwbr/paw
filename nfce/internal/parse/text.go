package parse

import (
	"strings"
	"unicode/utf8"
)

// Text normalizes whitespace by replacing special characters, trimming, and collapsing spaces.
func Text(s string) string {
	replacer := strings.NewReplacer(
		"\u00A0", " ", // NBSP
		"\u200B", "", // zero-width space
	)

	s = replacer.Replace(s)
	s = strings.TrimSpace(s)

	if s == "" {
		return ""
	}

	return strings.Join(strings.Fields(s), " ")
}

// StripNonNumericPrefix removes leading non-numeric characters before parsing.
func StripNonNumericPrefix(s string) string {
	s = Text(s)

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if (r >= '0' && r <= '9') || r == '-' {
			break
		}
		s = s[size:]
	}
	return s
}

var accentReplacer = strings.NewReplacer(
	"Á", "A", "À", "A", "Ã", "A", "Â", "A",
	"É", "E", "Ê", "E",
	"Í", "I",
	"Ó", "O", "Õ", "O", "Ô", "O",
	"Ú", "U",
	"Ç", "C",
)

// Normalize converts text for case-insensitive comparison by removing accents and converting to uppercase.
func Normalize(s string) string {
	return accentReplacer.Replace(strings.ToUpper(s))
}

// DigitsOnly returns s with all non-digit characters removed.
func DigitsOnly(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

// ContainsAny returns true if s contains any of the given terms.
func ContainsAny(s string, terms ...string) bool {
	for _, t := range terms {
		if strings.Contains(s, t) {
			return true
		}
	}
	return false
}

// BRNumber converts a Brazilian-formatted number string into a Go-parseable
// decimal: thousands separator "." is removed, decimal "," becomes ".".
// Example: "1.197,00" -> "1197.00", "12,90" -> "12.90".
func BRNumber(s string) string {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return s
}
