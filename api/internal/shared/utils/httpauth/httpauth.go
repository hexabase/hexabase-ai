package httpauth

import "strings"

const bearerPrefix = "bearer "

// HasBearerPrefix checks if the Authorization header has a "Bearer " prefix.
// It performs a case-insensitive check.
func HasBearerPrefix(header string) bool {
	if len(header) < len(bearerPrefix) {
		return false
	}
	return strings.ToLower(header[:len(bearerPrefix)]) == bearerPrefix
}

// TrimBearerPrefix removes the "Bearer " prefix from the Authorization header.
// It performs a case-insensitive check.
func TrimBearerPrefix(header string) string {
	if HasBearerPrefix(header) {
		return header[len(bearerPrefix):]
	}
	return header
} 