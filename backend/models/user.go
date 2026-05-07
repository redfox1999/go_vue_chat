package models

import "time"

type User struct {
	ID       int        `json:"id" db:"id"`
	Username string     `json:"username" db:"username"`
	Nickname string     `json:"nickname" db:"nickname"`
	Email    string     `json:"email" db:"email"`
	Password string     `json:"-" db:"password"`
	Birthday *time.Time `json:"birthday,omitempty" db:"birthday"`
	Sign     string     `json:"sign" db:"sign"`
	Status   int        `json:"status" db:"status"`
	CreateAt time.Time  `json:"create_at" db:"create_at"`
	UpdateAt time.Time  `json:"update_at" db:"update_at"`
}
