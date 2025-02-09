package cms

import (
	"github.com/emvi/shifu/pkg/storage"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Init(storage.NewFileStorage())
	os.Exit(m.Run())
}
