package content

// TemplateConfig is the configuration for a template.
type TemplateConfig struct {
	Name    string            `json:"-"`
	Label   string            `json:"label"`
	Content map[string]string `json:"content"`
}

// Positions returns the available positions for the template.
func (c *TemplateConfig) Positions() []string {
	positions := make([]string, 0, len(c.Content))

	for k := range c.Content {
		positions = append(positions, k)
	}

	return positions
}
