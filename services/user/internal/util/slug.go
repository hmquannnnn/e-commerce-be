package util

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	// Regex to match non-alphanumeric characters
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
)

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove Vietnamese accents
	s = removeAccents(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove all non-alphanumeric characters except hyphens
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")

	// Remove leading and trailing hyphens
	s = strings.Trim(s, "-")

	// Replace multiple consecutive hyphens with single hyphen
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	return s
}

// removeAccents removes accents from Vietnamese and other characters
func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)

	// Additional Vietnamese character replacements
	replacements := map[string]string{
		"đ": "d",
		"Đ": "d",
	}

	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// IsValidSlug checks if a slug is valid
func IsValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}

	// Must start and end with alphanumeric character
	if !isAlphanumeric(rune(slug[0])) || !isAlphanumeric(rune(slug[len(slug)-1])) {
		return false
	}

	// Can only contain lowercase alphanumeric characters and hyphens
	for _, char := range slug {
		if !isAlphanumeric(char) && char != '-' {
			return false
		}
	}

	// Cannot contain consecutive hyphens
	if strings.Contains(slug, "--") {
		return false
	}

	return true
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
