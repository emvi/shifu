package content

import (
	"encoding/json"
	"github.com/emvi/shifu/pkg/cfg"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

const (
	templateConfigDir = "admin/tpl"
)

// TemplateCfgCache is a cache for template configurations.
type TemplateCfgCache struct {
	templates map[string]TemplateConfig
	positions map[string]string
	list      []TemplateConfig
	m         sync.RWMutex
}

// NewTemplateCfgCache creates a new TemplateCfgCache.
func NewTemplateCfgCache() *TemplateCfgCache {
	cache := &TemplateCfgCache{
		templates: make(map[string]TemplateConfig),
		positions: make(map[string]string),
	}
	cache.Load()
	return cache
}

// Load loads the template cache from disk.
func (c *TemplateCfgCache) Load() {
	configDir := filepath.Join(cfg.Get().BaseDir, templateConfigDir)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		slog.Warn("Administration template configuration directory does not exist")
		return
	}

	templates := make(map[string]TemplateConfig)
	positions := make(map[string]string)
	list := make([]TemplateConfig, 0)

	if err := filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Contains(path, ".json") {
			name := strings.TrimSuffix(info.Name(), ".json")
			tpl := c.loadTemplate(path)

			if tpl == nil {
				return nil
			}

			tpl.Name = name
			templates[name] = *tpl
			list = append(list, *tpl)

			for k, v := range tpl.Content {
				positions[k] = v.Label
			}
		}

		return err
	}); err != nil {
		slog.Error("Error loading template configurations", "error", err)
		return
	}

	slices.SortFunc(list, func(a, b TemplateConfig) int {
		nameA := strings.ToLower(a.Label)

		if nameA == "" {
			nameA = a.Name
		}

		nameB := strings.ToLower(b.Label)

		if nameB == "" {
			nameB = b.Name
		}

		if nameA < nameB {
			return -1
		} else if nameA > nameB {
			return 1
		}

		return 0
	})
	c.m.Lock()
	defer c.m.Unlock()
	c.templates = templates
	c.positions = positions
	c.list = list
}

// List returns a list of all template configurations.
// The filter is optional.
func (c *TemplateCfgCache) List(filter []string) []TemplateConfig {
	c.m.RLock()
	defer c.m.RUnlock()

	if len(filter) == 0 {
		list := make([]TemplateConfig, len(c.list))
		copy(list, c.list)
		return list
	}

	list := make([]TemplateConfig, 0, len(c.list))

	for _, tpl := range c.list {
		if slices.Contains(filter, tpl.Name) {
			list = append(list, tpl)
		}
	}

	return list
}

// GetTemplate returns a template configuration by name.
func (c *TemplateCfgCache) GetTemplate(name string) (TemplateConfig, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	tpl, found := c.templates[name]
	return tpl, found
}

// GetPositions returns the template position mapping.
func (c *TemplateCfgCache) GetPositions() map[string]string {
	c.m.RLock()
	defer c.m.RUnlock()
	m := make(map[string]string)

	for k, v := range c.positions {
		m[k] = v
	}

	return m
}

func (c *TemplateCfgCache) loadTemplate(path string) *TemplateConfig {
	file, err := os.ReadFile(path)

	if err != nil {
		slog.Warn("TemplateConfig configuration not found", "error", err, "path", path)
		return nil
	}

	var tpl TemplateConfig

	if err := json.Unmarshal(file, &tpl); err != nil {
		slog.Error("Error unmarshalling template configuration", "error", err, "path", path)
		return nil
	}

	return &tpl
}
