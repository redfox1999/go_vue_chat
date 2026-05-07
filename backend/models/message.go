package models

import "time"

type Message struct {
	ID       int       `json:"id" db:"id"`
	RoomID   int       `json:"room_id" db:"room_id"`
	Sender   int       `json:"sender" db:"sender"`
	Notify   string    `json:"notify" db:"notify"`
	Message  string    `json:"message" db:"message"`
	SendTime time.Time `json:"send_time" db:"send_time"`
}
