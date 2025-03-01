package tpl

var (
	tpl *Cache
)

func init() {
	tpl = NewCache()
}

// Get returns the template cache.
func Get() *Cache {
	return tpl
}
