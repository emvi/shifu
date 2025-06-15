package static

import "embed"

//go:embed admin/tpl
var AdminTpl embed.FS

//go:embed admin/static
var AdminStatic embed.FS
