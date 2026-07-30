package ai

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const FaithfulnessMetricName = "numeric_tool_groundedness"

var numberTokenRe = regexp.MustCompile(`(?i)#?\d[\d,]*(?:\.\d+)?\+?(?:\s*[km])?%?`)

type FaithfulnessToolSource struct {
	ToolName string `json:"tool_name"`
	Response any    `json:"response,omitempty"`
}

type FaithfulnessClaim struct {
	Raw     string `json:"raw"`
	Value   string `json:"value"`
	Context string `json:"context"`
}

type FaithfulnessEval struct {
	Metric             string              `json:"metric"`
	Passed             bool                `json:"passed"`
	Score              float64             `json:"score"`
	ClaimsChecked      int                 `json:"claims_checked"`
	ClaimsSupported    int                 `json:"claims_supported"`
	ClaimsUnsupported  int                 `json:"claims_unsupported"`
	UnsupportedClaims  []FaithfulnessClaim `json:"unsupported_claims,omitempty"`
	ToolResponseCount  int                 `json:"tool_response_count"`
	ReferenceFactCount int                 `json:"reference_fact_count"`
	Reason             string              `json:"reason"`
}

type numericClaim struct {
	FaithfulnessClaim
	value      float64
	decimals   int
	multiplier float64
	plus       bool
}

type numericFact struct {
	value   float64
	integer bool
}

// EvaluateFaithfulness checks whether numeric claims in an answer are grounded
// in the tool responses used by the agent. It intentionally focuses on numeric
// facts because this product's AI answers are stats-heavy and prompt rules
// require retrieved numbers for ranks, stars, repos, scores, and contributions.
func EvaluateFaithfulness(answer string, sources []FaithfulnessToolSource) *FaithfulnessEval {
	claims := extractNumericClaims(answer)
	facts := collectReferenceFacts(sources)

	out := &FaithfulnessEval{
		Metric:             FaithfulnessMetricName,
		Passed:             true,
		Score:              1,
		ClaimsChecked:      len(claims),
		ToolResponseCount:  len(sources),
		ReferenceFactCount: len(facts),
		Reason:             "all numeric claims are grounded in tool responses",
	}
	if len(claims) == 0 {
		out.Reason = "no numeric claims to check"
		return out
	}
	if len(facts) == 0 {
		out.Passed = false
		out.Score = 0
		out.ClaimsUnsupported = len(claims)
		out.UnsupportedClaims = publicClaims(claims)
		out.Reason = "numeric claims were present but no tool facts were available"
		return out
	}

	for _, claim := range claims {
		if claimSupported(claim, facts) {
			out.ClaimsSupported++
			continue
		}
		out.UnsupportedClaims = append(out.UnsupportedClaims, claim.FaithfulnessClaim)
	}

	out.ClaimsUnsupported = len(out.UnsupportedClaims)
	out.Passed = out.ClaimsUnsupported == 0
	out.Score = roundScore(float64(out.ClaimsSupported) / float64(out.ClaimsChecked))
	if !out.Passed {
		out.Reason = "one or more numeric claims were not found in tool responses"
	}
	return out
}

func extractNumericClaims(s string) []numericClaim {
	matches := numberTokenRe.FindAllStringIndex(s, -1)
	claims := make([]numericClaim, 0, len(matches))
	for _, match := range matches {
		raw := s[match[0]:match[1]]
		if shouldIgnoreAnswerNumber(s, match[0], match[1], raw) {
			continue
		}
		value, decimals, multiplier, plus, ok := parseNumberToken(raw)
		if !ok {
			continue
		}
		claims = append(claims, numericClaim{
			FaithfulnessClaim: FaithfulnessClaim{
				Raw:     raw,
				Value:   formatFloat(value),
				Context: numberContext(s, match[0], match[1]),
			},
			value:      value,
			decimals:   decimals,
			multiplier: multiplier,
			plus:       plus,
		})
	}
	return claims
}

func collectReferenceFacts(sources []FaithfulnessToolSource) []numericFact {
	var facts []numericFact
	for _, source := range sources {
		collectFacts(source.Response, "response", &facts)
	}
	return facts
}

func collectFacts(v any, path string, facts *[]numericFact) {
	if ignoredFactPath(path) {
		return
	}
	switch val := v.(type) {
	case nil:
		return
	case map[string]any:
		for k, child := range val {
			collectFacts(child, path+"."+k, facts)
		}
	case []any:
		for _, child := range val {
			collectFacts(child, path, facts)
		}
	case json.Number:
		if f, err := val.Float64(); err == nil {
			*facts = append(*facts, numericFact{value: f, integer: isWhole(f)})
		}
	case float64:
		*facts = append(*facts, numericFact{value: val, integer: isWhole(val)})
	case float32:
		f := float64(val)
		*facts = append(*facts, numericFact{value: f, integer: isWhole(f)})
	case int:
		*facts = append(*facts, numericFact{value: float64(val), integer: true})
	case int64:
		*facts = append(*facts, numericFact{value: float64(val), integer: true})
	case int32:
		*facts = append(*facts, numericFact{value: float64(val), integer: true})
	case uint:
		*facts = append(*facts, numericFact{value: float64(val), integer: true})
	case uint64:
		*facts = append(*facts, numericFact{value: float64(val), integer: true})
	case uint32:
		*facts = append(*facts, numericFact{value: float64(val), integer: true})
	case string:
		for _, claim := range extractNumericClaims(val) {
			*facts = append(*facts, numericFact{value: claim.value, integer: isWhole(claim.value)})
		}
	}
}

func claimSupported(claim numericClaim, facts []numericFact) bool {
	for _, fact := range facts {
		if factSupportsClaim(fact, claim) {
			return true
		}
	}
	return false
}

func factSupportsClaim(fact numericFact, claim numericClaim) bool {
	if claim.plus {
		return fact.value >= claim.value
	}
	if fact.integer && claim.decimals == 0 && claim.multiplier == 1 {
		return fact.value == claim.value
	}
	return math.Abs(fact.value-claim.value) <= claimTolerance(claim)
}

func claimTolerance(claim numericClaim) float64 {
	if claim.decimals <= 0 {
		return 0.5 * claim.multiplier
	}
	return 0.5 * math.Pow10(-claim.decimals) * claim.multiplier
}

func parseNumberToken(raw string) (float64, int, float64, bool, bool) {
	token := strings.ToLower(strings.TrimSpace(raw))
	token = strings.TrimPrefix(token, "#")
	token = strings.ReplaceAll(token, ",", "")
	token = strings.ReplaceAll(token, " ", "")
	token = strings.TrimSuffix(token, "%")

	plus := strings.HasSuffix(token, "+")
	token = strings.TrimSuffix(token, "+")

	multiplier := 1.0
	if strings.HasSuffix(token, "k") {
		multiplier = 1000
		token = strings.TrimSuffix(token, "k")
	} else if strings.HasSuffix(token, "m") {
		multiplier = 1000000
		token = strings.TrimSuffix(token, "m")
	}

	decimals := 0
	if dot := strings.Index(token, "."); dot != -1 {
		decimals = len(token) - dot - 1
	}
	value, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return 0, 0, 1, false, false
	}
	return value * multiplier, decimals, multiplier, plus, true
}

func shouldIgnoreAnswerNumber(s string, start, end int, raw string) bool {
	if adjacentLetter(s, end) {
		return true
	}
	cleaned := strings.TrimPrefix(strings.TrimSpace(raw), "#")
	cleaned = strings.TrimSuffix(cleaned, "%")
	cleaned = strings.TrimSuffix(cleaned, "+")
	if len(cleaned) == 4 {
		if year, err := strconv.Atoi(cleaned); err == nil && year >= 1900 && year <= 2100 {
			return true
		}
	}
	context := strings.ToLower(numberContext(s, start, end))
	return strings.Contains(context, "http://") || strings.Contains(context, "https://")
}

func adjacentLetter(s string, index int) bool {
	if index >= len(s) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s[index:])
	return unicode.IsLetter(r)
}

func ignoredFactPath(path string) bool {
	path = strings.ToLower(path)
	ignored := []string{
		".id", "_id", "url", "avatar", "website", "email", "date", "time",
		"created", "updated", "synced", "period_start", "period_end", "hash",
	}
	for _, part := range ignored {
		if strings.Contains(path, part) {
			return true
		}
	}
	return false
}

func numberContext(s string, start, end int) string {
	left := start - 45
	if left < 0 {
		left = 0
	}
	right := end + 45
	if right > len(s) {
		right = len(s)
	}
	return strings.TrimSpace(s[left:right])
}

func publicClaims(claims []numericClaim) []FaithfulnessClaim {
	out := make([]FaithfulnessClaim, 0, len(claims))
	for _, claim := range claims {
		out = append(out, claim.FaithfulnessClaim)
	}
	return out
}

func isWhole(v float64) bool {
	return math.Abs(v-math.Round(v)) < 0.0000001
}

func formatFloat(v float64) string {
	if isWhole(v) {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func roundScore(v float64) float64 {
	return math.Round(v*100) / 100
}
