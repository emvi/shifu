package content

// TemplateConfig is the configuration for a template.
type TemplateConfig struct {
	Name    string                     `json:"-"`
	Label   string                     `json:"label"`
	Layout  bool                       `json:"layout"`
	Content map[string]TemplateContent `json:"content"`
	Copy    map[string]TemplateCopy    `json:"copy"`
	Data    map[string]TemplateData    `json:"data"`
}

// Positions returns the available positions for the template.
func (c *TemplateConfig) Positions() []string {
	positions := make([]string, 0, len(c.Content))

	for k := range c.Content {
		positions = append(positions, k)
	}

	return positions
}

// TemplateContent is the configuration for a template content area.
type TemplateContent struct {
	Label     string   `json:"label"`
	TplFilter []string `json:"tpl_filter"`
}

// TemplateCopy is the configuration for a template copy entry.
type TemplateCopy struct {
	Label   string            `json:"label"`
	Type    string            `json:"type"`
	Options map[string]string `json:"options"`
}

// TemplateData is the configuration for a template data entry.
type TemplateData struct {
	Label   string            `json:"label"`
	Type    string            `json:"type"`
	Options map[string]string `json:"options"`
}
