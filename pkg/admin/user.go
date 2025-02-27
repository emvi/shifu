package admin

import (
	"github.com/emvi/shifu/pkg/admin/model"
	"log/slog"
	"net/http"
)

func User(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, "user.html", struct {
		Admin         bool
		Self          *model.User
		WindowOptions WindowOptions
		User          []model.User
	}{
		Admin: isAdmin(r),
		Self:  getUser(r),
		WindowOptions: WindowOptions{
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

	if err := db.Select(&user, `SELECT * FROM "user" ORDER BY id`); err != nil {
		slog.Error("Error selecting user", "error", err)
	}

	return user
}
