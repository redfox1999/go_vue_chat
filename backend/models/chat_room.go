package models

import "time"

type ChatRoom struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Logo      string    `json:"logo" db:"logo"`
	Desc      string    `json:"desc" db:"desc"`
	OwnerID   int       `json:"owner_id" db:"owner_id"`
	Group     string    `json:"group" db:"group"`
	Status    int       `json:"status" db:"status"`
	CreateAt  time.Time `json:"create_at" db:"create_at"`
	UpdateAt  time.Time `json:"update_at" db:"update_at"`
}
