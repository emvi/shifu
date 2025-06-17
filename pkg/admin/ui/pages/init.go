package pages

import (
	"github.com/emvi/shifu/pkg/cms"
)

var (
	content *cms.CMS
)

// Init initializes the template cache.
func Init(cms *cms.CMS) {
	content = cms
}
