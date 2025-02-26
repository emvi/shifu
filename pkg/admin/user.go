package admin

import (
	"github.com/emvi/shifu/pkg/admin/model"
	"log/slog"
	"net/http"
)

func User(w http.ResponseWriter, _ *http.Request) {
	var user []model.User

	if err := db.Select(&user, `SELECT * FROM "user" ORDER BY id`); err != nil {
		slog.Error("Error selecting user ", "error", err)
		return
	}

	tpl.Execute(w, "user.html", struct {
		WindowOptions WindowOptions
		User          []model.User
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-user",
			TitleTpl:   "user-window-title",
			ContentTpl: "user-window-content",
			MinWidth:   500,
		},
		User: user,
	})
}
