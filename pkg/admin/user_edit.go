package admin

import (
	"database/sql"
	"errors"
	"github.com/emvi/shifu/pkg/admin/model"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func EditUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	email := strings.TrimSpace(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("name"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	user := new(model.User)

	if id != "" {
		if err := db.Get(user, `SELECT * FROM "user" WHERE id = ?`, id); err != nil {
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

	if !isAdmin(r) && userID != user.ID {
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

			if err := db.Get(&exists, `SELECT 1 FROM "user" WHERE email = ? AND id != ?`, email, user.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
				slog.Error("Error selecting user by email", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if exists {
				errs["email"] = "the email address is already used"
			}
		}

		if name == "" {
			errs["name"] = "the name address is required"
		}

		if password != passwordConfirm {
			errs["password_confirm"] = "the passwords do not match"
		} else if user.ID == 0 && password == "" {
			errs["password"] = "the password is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Execute(w, "user-edit-form.html", struct {
				User   *model.User
				Email  string
				Name   string
				Errors map[string]string
			}{
				User:   user,
				Email:  email,
				Name:   name,
				Errors: errs,
			})
			return
		}

		if user.ID == 0 {
			slog.Info("Creating new user")
			passwordSalt := GenRandomString(20)
			passwordHash := HashPassword(password + passwordSalt)

			if _, err := db.Exec(`INSERT INTO "user" (email, full_name, password, password_salt) VALUES (?, ?, ?, ?)`, email, name, passwordHash, passwordSalt); err != nil {
				slog.Error("Error inserting user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			if password != "" {
				slog.Info("Saving user with password")
				passwordSalt := GenRandomString(20)
				passwordHash := HashPassword(password + passwordSalt)

				if _, err := db.Exec(`UPDATE "user" SET email = ?, full_name = ?, password = ?, password_salt = ? WHERE id = ?`, email, name, passwordHash, passwordSalt, user.ID); err != nil {
					slog.Error("Error updating user", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				slog.Info("Saving user without password")

				if _, err := db.Exec(`UPDATE "user" SET email = ?, full_name = ? WHERE id = ?`, email, name, user.ID); err != nil {
					slog.Error("Error updating user", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		tpl.Execute(w, "user-list.html", struct {
			Admin bool
			Self  *model.User
			User  []model.User
		}{
			Admin: isAdmin(r),
			Self:  getUser(r),
			User:  listUser(),
		})
		return
	}

	tpl.Execute(w, "user-edit.html", struct {
		WindowOptions WindowOptions
		User          *model.User
		Email         string
		Name          string
		Errors        map[string]string
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-user-edit",
			TitleTpl:   "user-edit-window-title",
			ContentTpl: "user-edit-window-content",
			MinWidth:   350,
			Overlay:    true,
		},
		User:  user,
		Email: email,
		Name:  name,
	})
}
