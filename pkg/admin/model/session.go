package model

import "time"

type Session struct {
	ID      int
	UserID  int `db:"user_id"`
	Session string
	Expires time.Time
}
