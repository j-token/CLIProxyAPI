package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

type OAuthCustomModelSummary struct {
	hash  string
	count int
}

// SummarizeOAuthCustomModels summarizes OAuth custom model declarations per channel.
func SummarizeOAuthCustomModels(entries map[string][]config.OAuthCustomModel) map[string]OAuthCustomModelSummary {
	if len(entries) == 0 {
		return nil
	}
	out := make(map[string]OAuthCustomModelSummary, len(entries))
	for k, v := range entries {
		key := strings.ToLower(strings.TrimSpace(k))
		if key == "" {
			continue
		}
		out[key] = summarizeOAuthCustomModelList(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// DiffOAuthCustomModelChanges compares OAuth custom model maps.
func DiffOAuthCustomModelChanges(oldMap, newMap map[string][]config.OAuthCustomModel) ([]string, []string) {
	oldSummary := SummarizeOAuthCustomModels(oldMap)
	newSummary := SummarizeOAuthCustomModels(newMap)
	keys := make(map[string]struct{}, len(oldSummary)+len(newSummary))
	for k := range oldSummary {
		keys[k] = struct{}{}
	}
	for k := range newSummary {
		keys[k] = struct{}{}
	}
	changes := make([]string, 0, len(keys))
	affected := make([]string, 0, len(keys))
	for key := range keys {
		oldInfo, okOld := oldSummary[key]
		newInfo, okNew := newSummary[key]
		switch {
		case okOld && !okNew:
			changes = append(changes, fmt.Sprintf("oauth-custom-models[%s]: removed", key))
			affected = append(affected, key)
		case !okOld && okNew:
			changes = append(changes, fmt.Sprintf("oauth-custom-models[%s]: added (%d entries)", key, newInfo.count))
			affected = append(affected, key)
		case okOld && okNew && oldInfo.hash != newInfo.hash:
			changes = append(changes, fmt.Sprintf("oauth-custom-models[%s]: updated (%d -> %d entries)", key, oldInfo.count, newInfo.count))
			affected = append(affected, key)
		}
	}
	sort.Strings(changes)
	sort.Strings(affected)
	return changes, affected
}

func summarizeOAuthCustomModelList(list []config.OAuthCustomModel) OAuthCustomModelSummary {
	if len(list) == 0 {
		return OAuthCustomModelSummary{}
	}
	seen := make(map[string]struct{}, len(list))
	normalized := make([]string, 0, len(list))
	for _, entry := range list {
		name := strings.ToLower(strings.TrimSpace(entry.Name))
		if name == "" {
			continue
		}
		key := name
		if displayName := strings.TrimSpace(entry.DisplayName); displayName != "" {
			key += "|display-name=" + displayName
		}
		if template := strings.TrimSpace(entry.Template); template != "" {
			key += "|template=" + strings.ToLower(template)
		}
		if entry.MaxContextLength > 0 {
			key += "|max-context-length=" + strconv.Itoa(entry.MaxContextLength)
		}
		if entry.Thinking != nil {
			key += "|thinking=" + strings.ToLower(strings.Join(entry.Thinking.Levels, ","))
			key += "|thinking-range=" + strconv.Itoa(entry.Thinking.Min) + "-" + strconv.Itoa(entry.Thinking.Max)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	if len(normalized) == 0 {
		return OAuthCustomModelSummary{}
	}
	sort.Strings(normalized)
	sum := sha256.Sum256([]byte(strings.Join(normalized, "|")))
	return OAuthCustomModelSummary{
		hash:  hex.EncodeToString(sum[:]),
		count: len(normalized),
	}
}
