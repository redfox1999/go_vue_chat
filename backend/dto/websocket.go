package dto

import "encoding/json"

// Message represents a WebSocket message structure
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// JoinPayload represents the payload for join action
type JoinPayload struct {
	RoomID string `json:"room_id"`
	Token  string `json:"token"`
}

// LeavePayload represents the payload for leave action
type LeavePayload struct {
	RoomID string `json:"room_id"`
}

// ChatPayload represents the payload for chat action
type ChatPayload struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

// ChatBroadcastPayload represents the payload for chat broadcast
type ChatBroadcastPayload struct {
	RoomID   string `json:"room_id"`
	UserID   int    `json:"user_id"`
	Nickname string `json:"nickname"`
	Content  string `json:"content"`
}

// UserJoinPayload represents the payload for user_join action
type UserJoinPayload struct {
	RoomID   string `json:"room_id"`
	UserID   int    `json:"user_id"`
	Nickname string `json:"nickname"`
}

// UserLeavePayload represents the payload for user_leave action
type UserLeavePayload struct {
	RoomID   string `json:"room_id"`
	UserID   int    `json:"user_id"`
	Nickname string `json:"nickname"`
}

// JoinOKPayload represents the payload for join_ok action
type JoinOKPayload struct {
	RoomID string `json:"room_id"`
}

// LeaveOKPayload represents the payload for leave_ok action
type LeaveOKPayload struct{}

// ErrorPayload represents the payload for error actions
type ErrorPayload struct {
	Error string `json:"error"`
}
