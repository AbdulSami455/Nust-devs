package ai

import "encoding/json"

func truncateJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if len(b) > 8000 {
		return string(b[:8000]) + "...[truncated]", nil
	}
	return string(b), nil
}
