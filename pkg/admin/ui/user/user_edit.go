package user

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/model"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/util"
	"github.com/emvi/shifu/pkg/middleware"
)

// EditUser renders the user creation or change dialog.
func EditUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	email := strings.TrimSpace(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("name"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	user := new(model.User)

	if id != "" {
		if err := db.Get().Get(user, `SELECT * FROM "user" WHERE id = ?`, id); err != nil {
			slog.Error("Error selecting user", "error", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if email == "" {
			email = user.Email
		}

		if name == "" {
			name = user.FullName
		}
	}

	userID, _ := strconv.Atoi(id)
	isAdmin := middleware.IsAdmin(r)

	if !isAdmin && userID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if user.Email == "admin" {
			email = user.Email
			name = user.FullName
		}

		errs := make(map[string]string)

		if email == "" {
			errs["email"] = "the email address is required"
		} else {
			exists := false

			if err := db.Get().Get(&exists, `SELECT 1 FROM "user" WHERE email = ? AND id != ?`, email, user.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
				slog.Error("Error selecting user by email", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if exists {
				errs["email"] = "the email address is already used"
			}
		}

		if name == "" {
			errs["name"] = "the name is required"
		}

		if password != passwordConfirm {
			errs["password_confirm"] = "the passwords do not match"
		} else if user.ID == 0 && password == "" {
			errs["password"] = "the password is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "user-edit-form.html", struct {
				Lang   string
				User   *model.User
				Email  string
				Name   string
				Errors map[string]string
			}{
				Lang:   tpl.GetUILanguage(r),
				User:   user,
				Email:  email,
				Name:   name,
				Errors: errs,
			})
			return
		}

		if user.ID == 0 {
			slog.Info("Creating new user")
			passwordSalt := util.GenRandomString(20)
			passwordHash := util.HashPassword(password + passwordSalt)

			if _, err := db.Get().Exec(`INSERT INTO "user" (email, full_name, password, password_salt) VALUES (?, ?, ?, ?)`, email, name, passwordHash, passwordSalt); err != nil {
				slog.Error("Error inserting user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			if password != "" {
				slog.Info("Saving user with password")
				passwordSalt := util.GenRandomString(20)
				passwordHash := util.HashPassword(password + passwordSalt)

				if _, err := db.Get().Exec(`UPDATE "user" SET email = ?, full_name = ?, password = ?, password_salt = ? WHERE id = ?`, email, name, passwordHash, passwordSalt, user.ID); err != nil {
					slog.Error("Error updating user", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				slog.Info("Saving user without password")

				if _, err := db.Get().Exec(`UPDATE "user" SET email = ?, full_name = ? WHERE id = ?`, email, name, user.ID); err != nil {
					slog.Error("Error updating user", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		tpl.Get().Execute(w, "user-list.html", struct {
			Lang  string
			Admin bool
			Self  *model.User
			User  []model.User
		}{
			Lang:  tpl.GetUILanguage(r),
			Admin: isAdmin,
			Self:  middleware.GetUser(r),
			User:  listUser(),
		})
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "user-edit.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		User          *model.User
		Email         string
		Name          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-user-edit",
			TitleTpl:   "user-edit-window-title",
			ContentTpl: "user-edit-window-content",
			MinWidth:   350,
			Overlay:    true,
			Lang:       lang,
		},
		Lang:  lang,
		User:  user,
		Email: email,
		Name:  name,
	})
}
