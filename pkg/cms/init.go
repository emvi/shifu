package cms

import "github.com/emvi/shifu/pkg/storage"

var (
	store storage.Storage
)

// Init initializes the CMS.
func Init(s storage.Storage) {
	store = s
}
