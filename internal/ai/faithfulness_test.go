package ai

import "testing"

func TestEvaluateFaithfulnessPassesGroundedNumericClaims(t *testing.T) {
	eval := EvaluateFaithfulness(
		"Based on the profile, @sami has 12 public repos, 48 stars, and rank #3.",
		[]FaithfulnessToolSource{{
			ToolName: "get_developer_snapshot",
			Response: map[string]any{
				"developer": map[string]any{
					"github_username": "sami",
					"public_repos":    12,
					"total_stars":     48,
				},
				"rank": 3,
			},
		}},
	)

	if !eval.Passed {
		t.Fatalf("expected eval to pass: %#v", eval)
	}
	if eval.ClaimsChecked != 3 || eval.ClaimsSupported != 3 || eval.ClaimsUnsupported != 0 {
		t.Fatalf("unexpected claim counts: %#v", eval)
	}
}

func TestEvaluateFaithfulnessFlagsUnsupportedNumericClaims(t *testing.T) {
	eval := EvaluateFaithfulness(
		"Based on the profile, @sami has 12 public repos and 99 stars.",
		[]FaithfulnessToolSource{{
			ToolName: "get_developer_snapshot",
			Response: map[string]any{
				"developer": map[string]any{
					"public_repos": 12,
					"total_stars":  48,
				},
			},
		}},
	)

	if eval.Passed {
		t.Fatalf("expected eval to fail: %#v", eval)
	}
	if eval.ClaimsChecked != 2 || eval.ClaimsSupported != 1 || eval.ClaimsUnsupported != 1 {
		t.Fatalf("unexpected claim counts: %#v", eval)
	}
	if len(eval.UnsupportedClaims) != 1 || eval.UnsupportedClaims[0].Value != "99" {
		t.Fatalf("unexpected unsupported claims: %#v", eval.UnsupportedClaims)
	}
}

func TestEvaluateFaithfulnessAllowsRoundedScores(t *testing.T) {
	eval := EvaluateFaithfulness(
		"Based on the leaderboard, the developer has a 92.5 activity score.",
		[]FaithfulnessToolSource{{
			ToolName: "get_leaderboard_snapshot",
			Response: map[string]any{
				"leaders": []any{
					map[string]any{"activity_score": 92.47},
				},
			},
		}},
	)

	if !eval.Passed {
		t.Fatalf("expected rounded score to pass: %#v", eval)
	}
}

func TestEvaluateFaithfulnessAllowsPlusClaims(t *testing.T) {
	eval := EvaluateFaithfulness(
		"Based on contribution stats, they have 120+ code reviews.",
		[]FaithfulnessToolSource{{
			ToolName: "get_developer_contribution_stats",
			Response: map[string]any{
				"reviews": 126,
			},
		}},
	)

	if !eval.Passed {
		t.Fatalf("expected plus claim to pass: %#v", eval)
	}
}

func TestEvaluateFaithfulnessPassesWhenThereAreNoNumericClaims(t *testing.T) {
	eval := EvaluateFaithfulness(
		"I could not find a matching developer.",
		nil,
	)

	if !eval.Passed {
		t.Fatalf("expected no numeric claims to pass: %#v", eval)
	}
	if eval.ClaimsChecked != 0 {
		t.Fatalf("expected no checked claims: %#v", eval)
	}
}
