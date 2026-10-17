package content

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/cms"
	"github.com/stretchr/testify/assert"
)

func TestEditFormOverrideFields(t *testing.T) {
	element := &cms.Content{Tpl: "text", Copy: cms.Copy{"en": {"headline": "Reference"}}, Data: map[string]any{"color": "#ff0000"}}
	override := &cms.Content{Ref: "ref", Copy: cms.Copy{"en": {"headline": "Page"}}}
	w := httptest.NewRecorder()
	tpl.Get().Execute(w, "page-element-edit-form.html", struct {
		Language     string
		Lang         string
		Path         string
		ElementPath  string
		Element      *cms.Content
		Config       TemplateConfig
		Languages    []string
		Copy         map[string]any
		Data         map[string]any
		Override     bool
		ShowOverride bool
		OverrideCopy map[string]any
		OverrideData map[string]any
	}{
		Language:    "en",
		Lang:        "en",
		ElementPath: "content.1",
		Element:     element,
		Config: TemplateConfig{
			Copy: map[string]TemplateCopy{"headline": {Label: "Headline"}},
			Data: map[string]TemplateData{"color": {Label: "Color", Type: "color"}},
		},
		Languages:    []string{"en"},
		Copy:         getCopy(element, "copy."),
		Data:         getData(element, "data."),
		Override:     true,
		ShowOverride: true,
		OverrideCopy: getCopy(override, "override.copy."),
		OverrideData: getData(override, "override.data."),
	})
	body := w.Body.String()
	assert.Equal(t, 200, w.Code)
	assert.True(t, strings.Contains(body, `name="copy.en.headline" value="Reference"`), body)
	assert.True(t, strings.Contains(body, `name="override.copy.en.headline" value="Page"`))
	assert.True(t, strings.Contains(body, `name="data.color" value="#ff0000"`))
	assert.True(t, strings.Contains(body, `name="override.data.color" value=""`))
}
