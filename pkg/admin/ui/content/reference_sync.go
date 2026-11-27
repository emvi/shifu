package content

import (
	"errors"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/admin"
	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/cfg"
)

// SyncReferences registers all reference files found in the references directory in the database
// and removes references from the database that no longer have a file.
func SyncReferences() error {
	dir := filepath.Join(cfg.Get().BaseDir, admin.RefsDir)
	names := make([]string, 0)

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".json" {
			names = append(names, strings.ToLower(strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))))
		}

		return nil
	}); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	for _, name := range names {
		res, err := db.Get().Exec(`INSERT OR IGNORE INTO "reference" ("name") VALUES (?)`, name)

		if err != nil {
			return err
		}

		if n, _ := res.RowsAffected(); n > 0 {
			slog.Info("Registered reference", "name", name)
		}
	}

	var registered []string

	if err := db.Get().Select(&registered, `SELECT "name" FROM "reference"`); err != nil {
		return err
	}

	for _, name := range registered {
		if !slices.Contains(names, name) {
			if _, err := db.Get().Exec(`DELETE FROM "reference" WHERE "name" = ?`, name); err != nil {
				return err
			}

			slog.Info("Removed reference", "name", name)
		}
	}

	return nil
}
