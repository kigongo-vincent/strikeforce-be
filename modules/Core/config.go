package core

import (
	"strings"
)

// GetFrontendURL returns the production frontend base URL.
// Always returns https://www.strikeforcetalent.africa regardless of environment.
// The URL is trimmed of trailing slashes.
func GetFrontendURL() string {
	baseURL := "https://www.strikeforcetalent.africa"
	return strings.TrimSuffix(baseURL, "/")
}
