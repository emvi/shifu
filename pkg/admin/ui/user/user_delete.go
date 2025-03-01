package user

import (
	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/middleware"
	"github.com/emvi/shifu/pkg/admin/model"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"log/slog"
	"net/http"
)

// DeleteUser renders the user deletion dialog.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	isAdmin := middleware.IsAdmin(r)

	if !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	user := new(model.User)

	if err := db.Get().Get(user, `SELECT * FROM "user" WHERE id = ?`, id); err != nil {
		slog.Error("Error selecting user", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if user.Email == "admin" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method == http.MethodDelete {
		if _, err := db.Get().Exec(`DELETE FROM "user" WHERE id = ?`, user.ID); err != nil {
			slog.Error("Error deleting user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tpl.Get().Execute(w, "user-list.html", struct {
			Admin bool
			User  []model.User
		}{
			Admin: isAdmin,
			User:  listUser(),
		})
		return
	}

	tpl.Get().Execute(w, "user-delete.html", struct {
		WindowOptions ui.WindowOptions
		User          *model.User
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-user-delete",
			TitleTpl:   "user-delete-window-title",
			ContentTpl: "user-delete-window-content",
			Overlay:    true,
		},
		User: user,
	})
}
