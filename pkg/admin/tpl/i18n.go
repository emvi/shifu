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
		"media_window_title":               "Media",
		"media_tree_edit":                  "Edit Directory",
		"media_tree_add":                   "Add Directory",
		"media_tree_delete":                "Delete Directory",
		"media_files_upload_files":         "Upload Files",
		"media_files_move_files":           "Move Files",
		"media_files_delete_files":         "Delete Files",
		"media_files_select":               "Select",
		"media_files_preview":              "Preview",
		"media_files_filename":             "Filename",
		"media_files_size":                 "Size",
		"media_files_edit":                 "Edit",
		"media_files_rename_file":          "Rename File",
		"media_files_move_file":            "Move File",
		"media_files_delete_file":          "Delete File",
		"media_files_empty":                "The directory is empty.",
		"media_files_no_directory":         "Please select a directory.",
		"media_create_dir_window_title":    "Create Directory",
		"media_create_dir_name":            "Name",
		"media_create_dir_submit":          "Create Directory",
		"media_delete_dir_window_title":    "Delete Directory",
		"media_delete_dir_confirm":         "Are you sure you want to delete the directory \"%s\"?",
		"media_delete_dir_submit":          "Delete Directory",
		"media_edit_dir_window_title":      "Edit Directory",
		"media_edit_dir_name":              "Name",
		"media_edit_dir_submit":            "Save",
		"media_delete_file_window_title":   "Delete Files",
		"media_delete_file_confirm":        "Are you sure you want to delete the file \"%s\"?",
		"media_delete_file_confirm_all":    "Are you sure you want to delete the following files?",
		"media_delete_file_warn":           "Warning! You might break links by deleting the files.",
		"media_delete_file_submit":         "Delete",
		"media_edit_file_window_title":     "Rename File",
		"media_edit_file_warn":             "Warning! You might break links by renaming the file.",
		"media_edit_file_name":             "Name",
		"media_edit_file_submit":           "Save",
		"media_move_file_window_title":     "Move Files",
		"media_move_file_confirm":          "Select a target directory to move the file \"%s\".",
		"media_move_file_confirm_all":      "Select a target directory to move the following files.",
		"media_move_file_warn":             "Warning! You might break links by moving the files.",
		"media_move_file_submit":           "Move",
		"media_upload_file_window_title":   "Upload Files",
		"media_upload_file_files":          "Files",
		"media_upload_file_overwrite":      "Overwrite Files",
		"media_upload_file_existing_files": "The following files already exist. Please change the name before you upload them or select the checkbox to overwrite.",
		"media_upload_file_submit":         "Upload",
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
		"media_window_title":               "Medien",
		"media_tree_edit":                  "Ordner bearbeiten",
		"media_tree_add":                   "Ordner anlegen",
		"media_tree_delete":                "Ordner löschen",
		"media_files_upload_files":         "Dateien hochladen",
		"media_files_move_files":           "Dateien verschieben",
		"media_files_delete_files":         "Dateien löschen",
		"media_files_select":               "Auswählen",
		"media_files_preview":              "Vorschau",
		"media_files_filename":             "Dateiname",
		"media_files_size":                 "Größe",
		"media_files_edit":                 "Bearbeiten",
		"media_files_rename_file":          "Datei umbenennen",
		"media_files_move_file":            "Datei verschieben",
		"media_files_delete_file":          "Datei löschen",
		"media_files_empty":                "Der Ordner ist leer.",
		"media_files_no_directory":         "Bitte wähle einen Ordner aus.",
		"media_create_dir_window_title":    "Ordner erstellen",
		"media_create_dir_name":            "Name",
		"media_create_dir_submit":          "Erstellen",
		"media_delete_dir_window_title":    "Ordner löschen",
		"media_delete_dir_confirm":         "Bist du sicher, dass du den Ordner \"%s\" löschen möchtest?",
		"media_delete_dir_submit":          "Ordner löschen",
		"media_edit_dir_window_title":      "Ordner bearbeiten",
		"media_edit_dir_name":              "Name",
		"media_edit_dir_submit":            "Speichern",
		"media_delete_file_window_title":   "Dateien löschen",
		"media_delete_file_confirm":        "Bist du sicher, dass du die Datei \"%s\" löschen möchtest?",
		"media_delete_file_confirm_all":    "Bist du dir sicher, dass du die folgenden Dateien löschen möchtest?",
		"media_delete_file_warn":           "Achtung! Beim Löschen können Links kaputtgehen!",
		"media_delete_file_submit":         "Löschen",
		"media_edit_file_window_title":     "Datei umbenennen",
		"media_edit_file_warn":             "Achtung! Beim Bearbeiten können Links kaputtgehen!",
		"media_edit_file_name":             "Name",
		"media_edit_file_submit":           "Speichern",
		"media_move_file_window_title":     "Dateien verschieben",
		"media_move_file_confirm":          "Wähle den Zielordner, um die Datei \"%s\" zu verschieben.",
		"media_move_file_confirm_all":      "Wähle den Zielordner, um die folgenden Dateien zu verschieben.",
		"media_move_file_warn":             "Achtung! Beim Verschieben können Links kaputtgehen!",
		"media_move_file_submit":           "Verschieben",
		"media_upload_file_window_title":   "Dateien hochladen",
		"media_upload_file_files":          "Dateien",
		"media_upload_file_overwrite":      "Dateien überschreiben",
		"media_upload_file_existing_files": "Die folgenden Dateien existieren bereits. Bitte ändere den Namen bevor du sie hochlädst oder wähle die Checkbox zum Überschreiben aus.",
		"media_upload_file_submit":         "Hochladen",
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
