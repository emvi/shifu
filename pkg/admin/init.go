package admin

import template "github.com/emvi/shifu/pkg/admin/tpl"

var (
	tpl *template.Cache
)

func init() {
	tpl = template.NewCache()
}
