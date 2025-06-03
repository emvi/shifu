package content

var (
	tplCache *TemplateCache
)

// Init initializes the template cache.
func Init() {
	tplCache = NewTemplateCache()
}
