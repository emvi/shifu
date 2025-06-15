package model

type User struct {
	ID           int
	Email        string
	Password     string
	PasswordSalt string `db:"password_salt"`
	FullName     string `db:"full_name"`
}
