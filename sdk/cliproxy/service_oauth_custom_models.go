package cliproxy

import (
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/modelconfig"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// applyOAuthCustomModels appends the models declared under oauth-custom-models for
// the auth's channel that are missing from the provider model list. API-key
// credentials keep their own per-entry models[] configuration and are skipped.
func applyOAuthCustomModels(cfg *config.Config, provider, authKind string, models []*ModelInfo) []*ModelInfo {
	if cfg == nil || len(cfg.OAuthCustomModels) == 0 {
		return models
	}
	channel := coreauth.OAuthModelAliasChannel(provider, authKind)
	if channel == "" {
		return models
	}
	entries := cfg.OAuthCustomModels[channel]
	if len(entries) == 0 {
		return models
	}
	return applyOAuthCustomModelEntries(channel, entries, models)
}

func applyOAuthCustomModelEntries(channel string, entries []config.OAuthCustomModel, models []*ModelInfo) []*ModelInfo {
	existing := make(map[string]struct{}, len(models)+len(entries))
	out := make([]*ModelInfo, 0, len(models)+len(entries))
	for _, model := range models {
		if model == nil {
			continue
		}
		out = append(out, model)
		if id := strings.TrimSpace(model.ID); id != "" {
			existing[strings.ToLower(id)] = struct{}{}
		}
	}
	now := time.Now().Unix()
	for i := range entries {
		name := strings.TrimSpace(entries[i].Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := existing[key]; ok {
			// The catalog already serves this slug; the declaration is redundant.
			continue
		}
		info := buildOAuthCustomModelInfo(channel, entries[i], models, now)
		if info == nil {
			continue
		}
		existing[key] = struct{}{}
		out = append(out, info)
	}
	return out
}

func buildOAuthCustomModelInfo(channel string, entry config.OAuthCustomModel, models []*ModelInfo, created int64) *ModelInfo {
	name := strings.TrimSpace(entry.Name)
	if name == "" {
		return nil
	}
	templateID := strings.TrimSpace(entry.Template)
	template := resolveOAuthCustomModelTemplate(templateID, name, models)
	if template == nil && templateID != "" {
		log.Warnf("oauth-custom-models[%s]: template %q for %q not found; using channel defaults", channel, templateID, name)
	}

	var info *ModelInfo
	if template != nil {
		clone := *template
		info = &clone
		if info.Name != "" {
			info.Name = rewriteModelInfoName(info.Name, template.ID, name)
		}
	} else {
		ownedBy, modelType := oauthCustomModelIdentity(channel, models)
		info = &ModelInfo{OwnedBy: ownedBy, Type: modelType}
	}
	info.ID = name
	info.Object = "model"
	info.Created = created
	info.Description = ""
	info.UserDefined = true
	if displayName := strings.TrimSpace(entry.DisplayName); displayName != "" {
		info.DisplayName = displayName
	} else {
		info.DisplayName = name
	}
	if entry.MaxContextLength > 0 {
		info.ContextLength = entry.MaxContextLength
		info.MaxContextLength = entry.MaxContextLength
	}
	if entry.Thinking != nil {
		info.Thinking = modelconfig.NormalizeThinkingSupport(entry.Thinking)
	}
	return info
}

// resolveOAuthCustomModelTemplate picks the catalog model whose capability metadata
// the custom model inherits. An explicit template is looked up in the channel list
// first and then in the static catalog. Without one, an exact static catalog match
// wins, then the channel model sharing the longest name prefix.
func resolveOAuthCustomModelTemplate(templateID, name string, models []*ModelInfo) *ModelInfo {
	if templateID != "" {
		if model := findModelInfoByID(models, templateID); model != nil {
			return model
		}
		return registry.LookupStaticModelInfo(templateID)
	}
	if model := registry.LookupStaticModelInfo(name); model != nil {
		return model
	}
	return longestPrefixModelInfo(models, name)
}

func findModelInfoByID(models []*ModelInfo, id string) *ModelInfo {
	for _, model := range models {
		if model != nil && strings.EqualFold(strings.TrimSpace(model.ID), id) {
			return model
		}
	}
	return nil
}

func longestPrefixModelInfo(models []*ModelInfo, name string) *ModelInfo {
	target := strings.ToLower(name)
	var best *ModelInfo
	bestLen := 0
	for _, model := range models {
		if model == nil {
			continue
		}
		candidate := strings.ToLower(strings.TrimSpace(model.ID))
		if candidate == "" {
			continue
		}
		n := commonPrefixLen(target, candidate)
		if n > bestLen {
			best = model
			bestLen = n
		}
	}
	return best
}

func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// oauthCustomModelIdentity returns the owned_by/type pair for a channel when no
// template model is available to copy them from.
func oauthCustomModelIdentity(channel string, models []*ModelInfo) (string, string) {
	for _, model := range models {
		if model != nil && model.OwnedBy != "" && model.Type != "" {
			return model.OwnedBy, model.Type
		}
	}
	switch channel {
	case "claude":
		return "anthropic", "claude"
	case "codex":
		return "openai", "openai"
	case "kimi":
		return "moonshot", "kimi"
	case "vertex", "aistudio":
		return "google", "gemini"
	default:
		return channel, channel
	}
}
