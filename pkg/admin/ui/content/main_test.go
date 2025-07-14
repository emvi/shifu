package content

import (
	"github.com/emvi/shifu/pkg/cfg"
	"os"
	"testing"
)

func TestMain(t *testing.M) {
	cfg.Get().BaseDir = "."
	os.Exit(t.Run())
}
