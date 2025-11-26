package user

import (
	"log/slog"
	"net/http"

	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/model"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/middleware"
)

// User renders the user management dialog.
func User(w http.ResponseWriter, r *http.Request) {
	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "user.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Admin         bool
		Self          *model.User
		User          []model.User
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-user",
			TitleTpl:   "user-window-title",
			ContentTpl: "user-window-content",
			MinWidth:   500,
			Lang:       lang,
		},
		Lang:  lang,
		Admin: middleware.IsAdmin(r),
		Self:  middleware.GetUser(r),
		User:  listUser(),
	})
}

func listUser() []model.User {
	var user []model.User

	if err := db.Get().Select(&user, `SELECT * FROM "user" ORDER BY id`); err != nil {
		slog.Error("Error selecting user", "error", err)
	}

	return user
}
