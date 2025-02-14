package cms

import "net/http"

// Handler is a special handler invoked if specified in Content.
type Handler func(*CMS, Content, map[string]string, http.ResponseWriter, *http.Request)

// Copy is the translated copy for Content.
type Copy map[string]map[string]any

// Sitemap is the sitemap data for the Content.
type Sitemap struct {
	Priority string `json:"priority"`
}

// Experiment is an A/B testing experiment.
type Experiment struct {
	Name    string `json:"name"`
	Variant string `json:"variant"`
}

// Analytics is the analytics metadata for the Content.
type Analytics struct {
	Tags       map[string]string `json:"tags"`
	Experiment Experiment        `json:"experiment"`
}

// Content is a page or element for the CMS.
type Content struct {
	DisableCache bool                 `json:"disable_cache"`
	Path         map[string]string    `json:"path"`
	Sitemap      Sitemap              `json:"sitemap"`
	Header       map[string]string    `json:"header"`
	Handler      string               `json:"handler"`
	Analytics    Analytics            `json:"analytics"`
	Ref          string               `json:"ref"`
	Tpl          string               `json:"tpl"`
	Data         map[string]any       `json:"data"`
	Copy         Copy                 `json:"copy"`
	Content      map[string][]Content `json:"content"`

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
}

func (content *Content) copy() Content {
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
			c[k][i] = j.copy()
		}
	}

	experiments := make(map[string][]string)

	for k, v := range content.Experiments {
		experiments[k] = make([]string, len(v))

		for i, j := range v {
			experiments[k][i] = j
		}
	}

	return Content{
		Path: path,
		Sitemap: Sitemap{
			Priority: content.Sitemap.Priority,
		},
		Header:  header,
		Handler: content.Handler,
		Analytics: Analytics{
			Tags:       tags,
			Experiment: content.Analytics.Experiment,
		},
		Ref:           content.Ref,
		Tpl:           content.Tpl,
		Data:          data,
		Copy:          contentCopy,
		Content:       c,
		Language:      content.Language,
		CanonicalLink: content.CanonicalLink,
		Experiments:   experiments,
	}
}
