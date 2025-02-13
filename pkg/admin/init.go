package admin

var (
	tpl *cache
)

func init() {
	tpl = newCache()
}
