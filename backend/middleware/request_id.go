package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIdKey contextKey = "request_id"
	UserIdKey    contextKey = "user_id"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), RequestIdKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIdKey).(string); ok {
		return requestID
	}
	return ""
}

func WithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, RequestIdKey, uuid.New().String())
}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIdKey, userID)
}

func GetUserID(ctx context.Context) interface{} {
	return ctx.Value(UserIdKey)
}
