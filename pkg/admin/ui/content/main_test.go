package content

import (
	"os"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
)

func TestMain(t *testing.M) {
	cfg.Get().BaseDir = "."
	os.Exit(t.Run())
}
