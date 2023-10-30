package utils

import (
	"strings"
)

// FormatMessage trims message to fit in first line of commits message
// or only the first 100 characters.
func FormatMessage(message string) string {
	msg := strings.TrimSpace(message)
	max := len(msg)
	if max > 100 {
		max = 100
	}
	newline := strings.Index(msg, "\n")
	if newline > 0 && max > newline {
		max = newline
	}
	return msg[:max]
}
