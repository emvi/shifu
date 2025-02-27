package admin

import (
	"github.com/emvi/shifu/pkg/admin/model"
	"log/slog"
	"net/http"
)

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	user := new(model.User)

	if err := db.Get(user, `SELECT * FROM "user" WHERE id = ?`, id); err != nil {
		slog.Error("Error selecting user", "error", err)
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
		if _, err := db.Exec(`DELETE FROM "user" WHERE id = ?`, user.ID); err != nil {
			slog.Error("Error deleting user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tpl.Execute(w, "user-list.html", struct {
			Admin bool
			User  []model.User
		}{
			Admin: isAdmin(r),
			User:  listUser(),
		})
		return
	}

	tpl.Execute(w, "user-delete.html", struct {
		WindowOptions WindowOptions
		User          *model.User
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-user-delete",
			TitleTpl:   "user-delete-window-title",
			ContentTpl: "user-delete-window-content",
			Overlay:    true,
		},
		User: user,
	})
}
