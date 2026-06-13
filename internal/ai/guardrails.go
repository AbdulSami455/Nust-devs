package ai

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const MaxMessageLength = 500

var (
	ErrMessageEmpty   = errors.New("message cannot be empty")
	ErrMessageTooLong = errors.New("message too long (max 500 characters)")
	ErrInjection      = errors.New("invalid message content")
	ErrOffTopic       = errors.New("I can only answer questions about NUST developers, their GitHub stats, projects, leaderboards, and community metrics")
)

var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(previous|all|above|prior)\s+instructions?`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+`),
	regexp.MustCompile(`(?i)act\s+as\s+(a|an|the)\s+`),
	regexp.MustCompile(`(?i)(system|admin|root)\s+prompt`),
	regexp.MustCompile(`(?i)disregard\s+(your|the)\s+(rules|guidelines|instructions)`),
	regexp.MustCompile(`(?i)do\s+anything\s+now`),
	regexp.MustCompile(`(?i)jailbreak`),
	regexp.MustCompile(`(?i)<\s*(script|iframe|img\s+src|svg)`),
	regexp.MustCompile(`(?i)prompt\s+injection`),
}

var offTopicKeywords = []string{
	"password", "api key", "private key", "secret", "token",
	"drop table", "truncate", "delete from", "insert into", "update set",
}

// ValidateMessage sanitizes and validates a user chat message.
// Returns the cleaned message or an error.
func ValidateMessage(msg string) (string, error) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", ErrMessageEmpty
	}
	if utf8.RuneCountInString(msg) > MaxMessageLength {
		return "", ErrMessageTooLong
	}
	for _, p := range injectionPatterns {
		if p.MatchString(msg) {
			return "", ErrInjection
		}
	}
	lower := strings.ToLower(msg)
	for _, kw := range offTopicKeywords {
		if strings.Contains(lower, kw) {
			return "", ErrOffTopic
		}
	}
	return msg, nil
}

var (
	bearerRe = regexp.MustCompile(`Bearer\s+[A-Za-z0-9\-._~+/]+=*`)
	skRe     = regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`)
	ghpRe    = regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`)
)

// SanitizeOutput strips accidentally leaked secrets from LLM output.
func SanitizeOutput(s string) string {
	s = bearerRe.ReplaceAllString(s, "Bearer [redacted]")
	s = skRe.ReplaceAllString(s, "[redacted]")
	s = ghpRe.ReplaceAllString(s, "[redacted]")
	return s
}

var githubUsernameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,38}$`)

// ValidateUsername checks a GitHub username is syntactically valid.
func ValidateUsername(u string) bool {
	return githubUsernameRe.MatchString(u)
}
