package ui

// WindowOptions are the options for the admin UI windows.
type WindowOptions struct {
	ID         string
	TitleTpl   string
	ContentTpl string
	MinWidth   int
	Overlay    bool
	Lang       string
}
