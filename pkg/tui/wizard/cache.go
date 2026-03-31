package wizard

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

type stepCache struct {
	entries map[string][]Choice
}

func newStepCache() *stepCache {
	return &stepCache{entries: make(map[string][]Choice)}
}

func (c *stepCache) get(stepName string, deps map[string]any) ([]Choice, bool) {
	key := cacheKey(stepName, deps)
	choices, ok := c.entries[key]
	return choices, ok
}

func (c *stepCache) set(stepName string, deps map[string]any, choices []Choice) {
	key := cacheKey(stepName, deps)
	c.entries[key] = choices
}

func (c *stepCache) invalidate(stepName string) {
	prefix := stepName + ":"
	for key := range c.entries {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.entries, key)
		}
	}
}

func cacheKey(stepName string, deps map[string]any) string {
	h := sha256.New()
	keys := make([]string, 0, len(deps))
	for k := range deps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// Use JSON for lossless serialization — %v is ambiguous for slices/strings.
		v, _ := json.Marshal(deps[k])
		_, _ = fmt.Fprintf(h, "%s=%s;", k, v)
	}
	return fmt.Sprintf("%s:%x", stepName, h.Sum(nil)[:8])
}
