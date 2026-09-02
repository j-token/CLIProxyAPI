package cliproxy

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/config"
)

func TestApplyOAuthCustomModels_AppendsFromExplicitTemplate(t *testing.T) {
	cfg := &config.Config{
		OAuthCustomModels: map[string][]config.OAuthCustomModel{
			"codex": {
				{Name: "gpt-daybreak-blue-latest", DisplayName: "Daybreak Blue", Template: "gpt-5.6-sol"},
			},
		},
	}
	models := []*ModelInfo{
		{
			ID:                        "gpt-5.6-sol",
			Object:                    "model",
			OwnedBy:                   "openai",
			Type:                      "openai",
			DisplayName:               "GPT 5.6 Sol",
			Description:               "sol description",
			ContextLength:             372000,
			MaxCompletionTokens:       128000,
			SupportedInputModalities:  []string{"text", "image"},
			SupportedOutputModalities: []string{"text"},
			Thinking:                  &registry.ThinkingSupport{Levels: []string{"low", "medium", "high"}},
		},
	}

	out := applyOAuthCustomModels(cfg, "codex", "oauth", models)
	if len(out) != 2 {
		t.Fatalf("expected 2 models, got %d", len(out))
	}
	custom := out[1]
	if custom.ID != "gpt-daybreak-blue-latest" {
		t.Fatalf("expected custom model id, got %q", custom.ID)
	}
	if custom.DisplayName != "Daybreak Blue" {
		t.Fatalf("expected configured display name, got %q", custom.DisplayName)
	}
	if custom.OwnedBy != "openai" || custom.Type != "openai" {
		t.Fatalf("expected template identity to be copied, got %s/%s", custom.OwnedBy, custom.Type)
	}
	if custom.ContextLength != 372000 || custom.MaxCompletionTokens != 128000 {
		t.Fatalf("expected template limits to be copied, got %d/%d", custom.ContextLength, custom.MaxCompletionTokens)
	}
	if custom.Thinking == nil || len(custom.Thinking.Levels) != 3 {
		t.Fatalf("expected template thinking levels to be copied, got %+v", custom.Thinking)
	}
	if custom.Description != "" {
		t.Fatalf("expected template description to be dropped, got %q", custom.Description)
	}
	if !custom.UserDefined {
		t.Fatalf("expected custom model to be marked user-defined")
	}
	if out[0].ID != "gpt-5.6-sol" || out[0].DisplayName != "GPT 5.6 Sol" {
		t.Fatalf("expected template model to stay untouched, got %+v", out[0])
	}
}

func TestApplyOAuthCustomModels_SkipsExistingCatalogModel(t *testing.T) {
	cfg := &config.Config{
		OAuthCustomModels: map[string][]config.OAuthCustomModel{
			"claude": {
				{Name: "Claude-Fable-5", DisplayName: "Should Not Apply"},
			},
		},
	}
	models := []*ModelInfo{
		{ID: "claude-fable-5", DisplayName: "Claude Fable 5"},
	}

	out := applyOAuthCustomModels(cfg, "claude", "oauth", models)
	if len(out) != 1 {
		t.Fatalf("expected 1 model, got %d", len(out))
	}
	if out[0].DisplayName != "Claude Fable 5" {
		t.Fatalf("expected catalog model to win, got %q", out[0].DisplayName)
	}
}

func TestApplyOAuthCustomModels_IgnoresAPIKeyCredentials(t *testing.T) {
	cfg := &config.Config{
		OAuthCustomModels: map[string][]config.OAuthCustomModel{
			"claude": {{Name: "claude-fable-5-1"}},
		},
	}
	models := []*ModelInfo{{ID: "claude-fable-5"}}

	out := applyOAuthCustomModels(cfg, "claude", "apikey", models)
	if len(out) != 1 {
		t.Fatalf("expected api-key credentials to be skipped, got %d models", len(out))
	}
}

func TestApplyOAuthCustomModels_PrefixTemplateWithOverrides(t *testing.T) {
	cfg := &config.Config{
		OAuthCustomModels: map[string][]config.OAuthCustomModel{
			"claude": {
				{
					Name:             "claude-fable-5-1-test-only",
					MaxContextLength: 1000000,
					Thinking:         &registry.ThinkingSupport{Levels: []string{"low", "high", "none"}},
				},
			},
		},
	}
	models := []*ModelInfo{
		{ID: "claude-opus-5", OwnedBy: "anthropic", Type: "claude", ContextLength: 200000},
		{ID: "claude-fable-5", OwnedBy: "anthropic", Type: "claude", ContextLength: 500000, MaxCompletionTokens: 128000},
	}

	out := applyOAuthCustomModels(cfg, "claude", "oauth", models)
	if len(out) != 3 {
		t.Fatalf("expected 3 models, got %d", len(out))
	}
	custom := out[2]
	if custom.MaxCompletionTokens != 128000 {
		t.Fatalf("expected longest-prefix template (claude-fable-5) to be used, got %+v", custom)
	}
	if custom.ContextLength != 1000000 || custom.MaxContextLength != 1000000 {
		t.Fatalf("expected max-context-length override, got %d/%d", custom.ContextLength, custom.MaxContextLength)
	}
	if custom.DisplayName != "claude-fable-5-1-test-only" {
		t.Fatalf("expected display name to default to the model name, got %q", custom.DisplayName)
	}
	if custom.Thinking == nil || len(custom.Thinking.Levels) != 3 || !custom.Thinking.ZeroAllowed {
		t.Fatalf("expected normalized thinking override, got %+v", custom.Thinking)
	}
}

func TestApplyOAuthCustomModels_WithoutTemplateUsesChannelIdentity(t *testing.T) {
	cfg := &config.Config{
		OAuthCustomModels: map[string][]config.OAuthCustomModel{
			"codex": {{Name: "zzz-unrelated-slug"}},
		},
	}

	out := applyOAuthCustomModels(cfg, "codex", "oauth", nil)
	if len(out) != 1 {
		t.Fatalf("expected 1 model, got %d", len(out))
	}
	if out[0].OwnedBy != "openai" || out[0].Type != "openai" || out[0].Object != "model" {
		t.Fatalf("expected channel identity defaults, got %+v", out[0])
	}
}

func TestApplyOAuthCustomModels_AliasCanReferenceCustomModel(t *testing.T) {
	cfg := &config.Config{
		OAuthCustomModels: map[string][]config.OAuthCustomModel{
			"codex": {{Name: "gpt-daybreak-blue-latest", Template: "gpt-5.6-sol"}},
		},
		OAuthModelAlias: map[string][]config.OAuthModelAlias{
			"codex": {{Name: "gpt-daybreak-blue-latest", Alias: "daybreak", Fork: true}},
		},
	}
	models := []*ModelInfo{{ID: "gpt-5.6-sol", OwnedBy: "openai", Type: "openai"}}

	out := applyOAuthCustomModels(cfg, "codex", "oauth", models)
	out = applyOAuthModelAlias(cfg, "codex", "oauth", out)
	if len(out) != 3 {
		t.Fatalf("expected 3 models, got %d", len(out))
	}
	if out[1].ID != "gpt-daybreak-blue-latest" || out[2].ID != "daybreak" {
		t.Fatalf("expected custom model and its alias, got %q and %q", out[1].ID, out[2].ID)
	}
}
