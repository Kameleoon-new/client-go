package utils

import "strings"

const (
	secretHiddenChar   = "*"
	secretVisibleCount = 4
)

func Secret(secret string) string {
	length := len(secret)
	if length <= secretVisibleCount {
		return strings.Repeat(secretHiddenChar, length)
	}
	visibleFrom := length - secretVisibleCount
	if visibleFrom < secretVisibleCount {
		visibleFrom = secretVisibleCount
	}
	return strings.Repeat(secretHiddenChar, secretVisibleCount) + secret[visibleFrom:]
}
