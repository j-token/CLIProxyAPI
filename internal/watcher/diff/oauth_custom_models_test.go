package diff

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestDiffOAuthCustomModelChanges(t *testing.T) {
	oldMap := map[string][]config.OAuthCustomModel{
		"codex":  {{Name: "gpt-daybreak-blue-latest", Template: "gpt-5.6-sol"}},
		"claude": {{Name: "claude-fable-5-1"}},
	}
	newMap := map[string][]config.OAuthCustomModel{
		"codex": {{Name: "gpt-daybreak-blue-latest", Template: "gpt-5.6-terra"}},
		"kimi":  {{Name: "kimi-k3"}},
	}

	changes, affected := DiffOAuthCustomModelChanges(oldMap, newMap)
	wantChanges := []string{
		"oauth-custom-models[claude]: removed",
		"oauth-custom-models[codex]: updated (1 -> 1 entries)",
		"oauth-custom-models[kimi]: added (1 entries)",
	}
	if len(changes) != len(wantChanges) {
		t.Fatalf("expected %d changes, got %v", len(wantChanges), changes)
	}
	for i := range wantChanges {
		if changes[i] != wantChanges[i] {
			t.Fatalf("change %d: expected %q, got %q", i, wantChanges[i], changes[i])
		}
	}
	if len(affected) != 3 || affected[0] != "claude" || affected[1] != "codex" || affected[2] != "kimi" {
		t.Fatalf("unexpected affected channels: %v", affected)
	}
}

func TestDiffOAuthCustomModelChanges_NoChange(t *testing.T) {
	entries := map[string][]config.OAuthCustomModel{
		"codex": {{Name: "gpt-daybreak-blue-latest", DisplayName: "Daybreak Blue"}},
	}
	changes, affected := DiffOAuthCustomModelChanges(entries, entries)
	if len(changes) != 0 || len(affected) != 0 {
		t.Fatalf("expected no changes, got %v / %v", changes, affected)
	}
}
