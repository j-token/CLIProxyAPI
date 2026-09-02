package config

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
)

func TestSanitizeOAuthCustomModels_NormalizesAndDedupes(t *testing.T) {
	cfg := &Config{
		OAuthCustomModels: map[string][]OAuthCustomModel{
			" Codex ": {
				{Name: " gpt-daybreak-blue-latest ", DisplayName: " Daybreak Blue ", Template: " gpt-5.6-sol "},
				{Name: "GPT-DAYBREAK-BLUE-LATEST", DisplayName: "duplicate"},
				{Name: "   "},
			},
			"claude": {
				{Name: "claude-fable-5-1", MaxContextLength: 1000000, Thinking: &registry.ThinkingSupport{Levels: []string{"low", "max"}}},
			},
			"": {{Name: "ignored"}},
		},
	}

	cfg.SanitizeOAuthCustomModels()

	if len(cfg.OAuthCustomModels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(cfg.OAuthCustomModels))
	}
	codex := cfg.OAuthCustomModels["codex"]
	if len(codex) != 1 {
		t.Fatalf("expected 1 codex entry, got %d", len(codex))
	}
	if codex[0].Name != "gpt-daybreak-blue-latest" || codex[0].DisplayName != "Daybreak Blue" || codex[0].Template != "gpt-5.6-sol" {
		t.Fatalf("unexpected codex entry: %+v", codex[0])
	}
	claude := cfg.OAuthCustomModels["claude"]
	if len(claude) != 1 || claude[0].MaxContextLength != 1000000 || claude[0].Thinking == nil || len(claude[0].Thinking.Levels) != 2 {
		t.Fatalf("unexpected claude entry: %+v", claude)
	}
}

func TestSanitizeOAuthCustomModels_DropsEmptyChannels(t *testing.T) {
	cfg := &Config{
		OAuthCustomModels: map[string][]OAuthCustomModel{
			"codex": {{Name: ""}},
		},
	}
	cfg.SanitizeOAuthCustomModels()
	if len(cfg.OAuthCustomModels) != 0 {
		t.Fatalf("expected empty map, got %+v", cfg.OAuthCustomModels)
	}
}
