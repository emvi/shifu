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

// User renders the user management dialog.
func User(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "user.html", struct {
		Admin         bool
		Self          *model.User
		WindowOptions ui.WindowOptions
		User          []model.User
	}{
		Admin: middleware.IsAdmin(r),
		Self:  middleware.GetUser(r),
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-user",
			TitleTpl:   "user-window-title",
			ContentTpl: "user-window-content",
			MinWidth:   500,
		},
		User: listUser(),
	})
}

func listUser() []model.User {
	var user []model.User

	if err := db.Get().Select(&user, `SELECT * FROM "user" ORDER BY id`); err != nil {
		slog.Error("Error selecting user", "error", err)
	}

	return user
}
