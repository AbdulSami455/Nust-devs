package githubutil

import (
	"errors"
	"regexp"
	"strings"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$|^[a-zA-Z0-9]$`)

var ErrInvalidUsername = errors.New("invalid github username format")

func NormalizeUsername(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if u == "" {
		return "", ErrInvalidUsername
	}
	// Strip @ prefix if pasted from URL
	u = strings.TrimPrefix(u, "@")
	u = strings.TrimSpace(u)
	if strings.Contains(u, "/") {
		parts := strings.Split(u, "/")
		u = parts[len(parts)-1]
	}
	u = strings.ToLower(u)
	if !usernamePattern.MatchString(u) {
		return "", ErrInvalidUsername
	}
	return u, nil
}
