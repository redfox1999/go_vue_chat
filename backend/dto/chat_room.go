package dto

import "time"

type CreateChatRoomRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=100"`
	Logo    string `json:"logo,omitempty"`
	Desc    string `json:"desc,omitempty" validate:"max=500"`
	Group   string `json:"group,omitempty"`
	OwnerID int    `json:"owner_id" validate:"required,min=1"`
}

type UpdateChatRoomRequest struct {
	Name    *string `json:"name"`
	Logo    *string `json:"logo"`
	Desc    *string `json:"desc"`
	Group   *string `json:"group"`
	Status  *int    `json:"status"`
	OwnerID *int    `json:"owner_id"`
}

type ChatRoomResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Logo      string    `json:"logo"`
	Desc      string    `json:"desc"`
	OwnerID   int       `json:"owner_id"`
	Group     string    `json:"group"`
	Status    int       `json:"status"`
	CreateAt  time.Time `json:"create_at"`
	UpdateAt  time.Time `json:"update_at"`
	ClientNum int       `json:"client_num"`
}

type ChatRoomListResponse struct {
	Data       []ChatRoomResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type MessageResponse struct {
	ID       int       `json:"id"`
	RoomID   int       `json:"room_id"`
	Sender   int       `json:"sender"`
	Nickname string    `json:"nickname"`
	Notify   string    `json:"notify"`
	Message  string    `json:"message"`
	SendTime time.Time `json:"send_time"`
}

type MessageListResponse struct {
	Data       []MessageResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}
