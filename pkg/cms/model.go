package cms

import "net/http"

// Handler is a special handler invoked if specified in Content.
type Handler func(*CMS, Content, map[string]string, http.ResponseWriter, *http.Request)

// Copy is the translated copy for Content.
type Copy map[string]map[string]any

// Sitemap is the sitemap data for the Content.
type Sitemap struct {
	Priority string `json:"priority,omitempty"`
}

// Experiment is an A/B testing experiment.
type Experiment struct {
	Name    string `json:"name"`
	Variant string `json:"variant"`
}

// Analytics is the analytics metadata for the Content.
type Analytics struct {
	Tags       map[string]string `json:"tags,omitempty"`
	Experiment Experiment        `json:"experiment,omitempty"`
}

// Content is a page or element for the CMS.
type Content struct {
	DisplayName  string               `json:"display_name"`
	DisableCache bool                 `json:"disable_cache,omitempty"`
	Path         map[string]string    `json:"path,omitempty"`
	Sitemap      Sitemap              `json:"sitemap"`
	Header       map[string]string    `json:"header,omitempty"`
	Handler      string               `json:"handler,omitempty"`
	Analytics    Analytics            `json:"analytics"`
	Ref          string               `json:"ref,omitempty"`
	Tpl          string               `json:"tpl,omitempty"`
	Data         map[string]any       `json:"data,omitempty"`
	Copy         Copy                 `json:"copy,omitempty"`
	Content      map[string][]Content `json:"content,omitempty"`

	// File is the path for the current content file.
	File string `json:"-"`

	// Request is the HTTP request.
	Request *http.Request `json:"-"`

	// Language is extracted and set from Path automatically.
	Language string `json:"-"`

	// CanonicalLink is set automatically using the configured hostname and Path.
	CanonicalLink string `json:"-"`

	// Experiments is a list of A/B experiments extracted from the content (name -> variants).
	Experiments map[string][]string `json:"-"`

	// SelectedExperiments is a list of selected A/B experiments from the Experiments list.
	SelectedExperiments map[string]string `json:"-"`

	// SelectedPageExperiment is an experiment from the page experiments list, redirecting if the visitor is on the wrong page.
	SelectedPageExperiment string `json:"-"`

	// Position is the element position path in JSON.
	Position string `json:"-"`
}

// Clone returns a copy of the Content.
func (content *Content) Clone() Content {
	path := make(map[string]string)

	for k, v := range content.Path {
		path[k] = v
	}

	header := make(map[string]string)

	for k, v := range content.Header {
		header[k] = v
	}

	tags := make(map[string]string)

	for k, v := range content.Analytics.Tags {
		tags[k] = v
	}

	data := make(map[string]any)

	for k, v := range content.Data {
		data[k] = v
	}

	contentCopy := make(Copy)

	for k, v := range content.Copy {
		contentCopy[k] = make(map[string]any)

		for i, j := range v {
			contentCopy[k][i] = j
		}
	}

	c := make(map[string][]Content)

	for k, v := range content.Content {
		c[k] = make([]Content, len(v))

		for i, j := range v {
			c[k][i] = j.Clone()
		}
	}

	experiments := make(map[string][]string)

	for k, v := range content.Experiments {
		experiments[k] = make([]string, len(v))

		for i, j := range v {
			experiments[k][i] = j
		}
	}

	selectedExperiments := make(map[string]string)

	for k, v := range content.SelectedExperiments {
		selectedExperiments[k] = v
	}

	return Content{
		DisplayName:  content.DisplayName,
		DisableCache: content.DisableCache,
		Path:         path,
		Sitemap: Sitemap{
			Priority: content.Sitemap.Priority,
		},
		Header:  header,
		Handler: content.Handler,
		Analytics: Analytics{
			Tags:       tags,
			Experiment: content.Analytics.Experiment,
		},
		Ref:                    content.Ref,
		Tpl:                    content.Tpl,
		Data:                   data,
		Copy:                   contentCopy,
		Content:                c,
		Request:                content.Request,
		Language:               content.Language,
		CanonicalLink:          content.CanonicalLink,
		Experiments:            experiments,
		SelectedExperiments:    selectedExperiments,
		SelectedPageExperiment: content.SelectedPageExperiment,
		Position:               content.Position,
	}
}
