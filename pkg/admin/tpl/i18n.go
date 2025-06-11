package tpl

import (
	"github.com/emvi/shifu/pkg/cms"
	"net/http"
)

var i18n = map[string]map[string]string{
	"en": {
		// window
		"close_window": "Close window",
		"cancel":       "Cancel",

		// login
		"login_meta_title":       "Shifu Login",
		"login_meta_description": "Sign in to Shifu CMS.",
		"login_headline":         "Login",
		"login_copyright":        "All rights reserved.",
		"login_website_link":     "Visit Website",
		"login_form_email":       "Email",
		"login_form_password":    "Password",
		"login_form_stay":        "Keep me logged in",
		"login_form_submit":      "Sign in",

		// toolbar
		"edit_page": "Edit page",
		"pages":     "Pages",
		"media":     "Media",
		"database":  "Database",
		"user":      "User",
		"sign_out":  "Sign out",

		// user management
		"user_window_title":              "User",
		"user_add":                       "Create User",
		"user_delete_window_title":       "Delete User",
		"user_delete_confirm":            "Are you sure you want to delete the user \"%s\"?",
		"user_delete_submit":             "Delete User",
		"user_edit_window_title":         "Edit User",
		"user_create_window_title":       "Create User",
		"user_edit_form_email":           "Email",
		"user_edit_form_name":            "Name",
		"user_edit_form_password":        "Password",
		"user_edit_form_repeat_password": "Repeat Password",
		"user_edit_form_submit":          "Save",
		"user_table_id":                  "ID",
		"user_table_email":               "Email",
		"user_table_name":                "Name",
		"user_table_edit":                "Edit",
		"user_table_tooltip_edit":        "Edit User",
		"user_table_tooltip_delete":      "Delete User",

		// database
		"db_window_title": "Database",
		"db_wip":          "Work in progress.",

		// media
		"media_window_title":       "Media",
		"media_tree_edit":          "Edit Directory",
		"media_tree_add":           "Add Directory",
		"media_tree_delete":        "Delete Directory",
		"media_files_upload_files": "Upload Files",
		"media_files_move_files":   "Move Files",
		"media_files_delete_files": "Delete Files",
		"media_files_select":       "Select",
		"media_files_preview":      "Preview",
		"media_files_filename":     "Filename",
		"media_files_size":         "Size",
		"media_files_edit":         "Edit",
		"media_files_rename_file":  "Rename File",
		"media_files_move_file":    "Move File",
		"media_files_delete_file":  "Delete File",
		"media_files_empty":        "The directory is empty.",
		"media_files_no_directory": "Please select a directory.",
	},
	"de": {
		// window
		"close_window": "Fenster schließen",
		"cancel":       "Abbrechen",

		// login
		"login_meta_title":       "Shifu Anmeldung",
		"login_meta_description": "Bei Shifu CMS anmelden.",
		"login_headline":         "Anmeldung",
		"login_copyright":        "Alle Rechte vorbehalten.",
		"login_website_link":     "Website besuchen",
		"login_form_email":       "Email",
		"login_form_password":    "Passwort",
		"login_form_stay":        "Angemeldet bleiben",
		"login_form_submit":      "Anmelden",

		// toolbar
		"edit_page": "Seite bearbeiten",
		"pages":     "Seiten",
		"media":     "Medien",
		"database":  "Datenbank",
		"user":      "Nutzer",
		"sign_out":  "Abmelden",

		// user management
		"user_window_title":              "Nutzerverwaltung",
		"user_add":                       "Benutzer hinzufügen",
		"user_delete_window_title":       "Nutzer löschen",
		"user_delete_confirm":            "Bist du sicher, dass du den Benutzer \"%s\" löschen möchtest?",
		"user_delete_submit":             "Nutzer löschen",
		"user_edit_window_title":         "Nutzer bearbeiten",
		"user_create_window_title":       "Nutzer erstellen",
		"user_edit_form_email":           "Email",
		"user_edit_form_name":            "Name",
		"user_edit_form_password":        "Passwort",
		"user_edit_form_repeat_password": "Passwort wiederholen",
		"user_edit_form_submit":          "Speichern",
		"user_table_id":                  "ID",
		"user_table_email":               "Email",
		"user_table_name":                "Name",
		"user_table_edit":                "Bearbeiten",
		"user_table_tooltip_edit":        "Nutzer bearbeiten",
		"user_table_tooltip_delete":      "Nutzer löschen",

		// database
		"db_window_title": "Datenbank",
		"db_wip":          "Nicht verfügbar.",

		// media
		"media_window_title":       "Medien",
		"media_tree_edit":          "Ordner bearbeiten",
		"media_tree_add":           "Ordner anlegen",
		"media_tree_delete":        "Ordner löschen",
		"media_files_upload_files": "Dateien hochladen",
		"media_files_move_files":   "Dateien verschieben",
		"media_files_delete_files": "Dateien löschen",
		"media_files_select":       "Auswählen",
		"media_files_preview":      "Vorschau",
		"media_files_filename":     "Dateiname",
		"media_files_size":         "Größe",
		"media_files_edit":         "Bearbeiten",
		"media_files_rename_file":  "Datei umbenennen",
		"media_files_move_file":    "Datei verschieben",
		"media_files_delete_file":  "Datei löschen",
		"media_files_empty":        "Der Ordner ist leer.",
		"media_files_no_directory": "Bitte wähle einen Ordner aus.",
	},
}

func getTranslation(language, key string) string {
	kv, found := i18n[language]

	if found {
		return kv[key]
	}

	return i18n["en"][key]
}

// GetLanguage returns the accepted language or "en" by default.
func GetLanguage(r *http.Request) string {
	languages := cms.GetAcceptedLanguages(r)

	if len(languages) == 0 {
		return "en"
	}

	return languages[0]
}
