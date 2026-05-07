package dto

import "time"

type CreateUserRequest struct {
	Username string     `json:"username" validate:"required,min=3,max=50"`
	Nickname string     `json:"nickname" validate:"max=100"`
	Email    string     `json:"email" validate:"required,email"`
	Password string     `json:"password" validate:"required,min=6,max=100"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Sign     string     `json:"sign,omitempty" validate:"max=500"`
	Status   int        `json:"status,omitempty"`
}

type UpdateUserRequest struct {
	Username *string    `json:"username"`
	Nickname *string    `json:"nickname"`
	Email    *string    `json:"email"`
	Password *string    `json:"password"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Sign     *string    `json:"sign"`
	Status   *int       `json:"status"`
}

type UserResponse struct {
	ID       int        `json:"id"`
	Username string     `json:"username"`
	Nickname string     `json:"nickname"`
	Email    string     `json:"email"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Sign     string     `json:"sign"`
	Status   int        `json:"status"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
}

type PaginationResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
