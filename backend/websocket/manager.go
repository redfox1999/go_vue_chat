package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"backend/models"
	"backend/repository"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize,
	WriteBufferSize: maxMessageSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	manager  *Manager
	send     chan []byte
	clientID string

	roomId   string
	userId   int
	nickName string

	closeOnce sync.Once // 确保关闭逻辑只执行一次
}

type RoomToken struct {
	OldToken string
	NewToken string
	ExpireAt time.Time
}

type Manager struct {
	clients      map[string]*Client
	rooms        map[string]map[string]*Client
	roomsToken   map[string]*RoomToken
	register     chan *Client
	unregister   chan *Client
	broadcast    chan []byte
	mu           sync.RWMutex
	logger       zerolog.Logger
	userRepo     repository.UserRepository
	chatRoomRepo repository.ChatRoomRepository
	messageRepo  repository.MessageRepository
}

func NewManager(logger zerolog.Logger, userRepo repository.UserRepository, chatRoomRepo repository.ChatRoomRepository, messageRepo repository.MessageRepository) *Manager {
	m := &Manager{
		clients:      make(map[string]*Client),
		rooms:        make(map[string]map[string]*Client),
		roomsToken:   make(map[string]*RoomToken),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan []byte),
		logger:       logger,
		userRepo:     userRepo,
		chatRoomRepo: chatRoomRepo,
		messageRepo:  messageRepo,
	}

	// 从数据库加载聊天室
	m.loadRoomsFromDB()

	return m
}

func (m *Manager) Run() {
	// 启动定时刷新 token 的 goroutine
	go m.startTokenRefreshTimer()

	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.clientID] = client
			var roomClientCount int
			if client.roomId != "" {
				if m.rooms[client.roomId] == nil {
					m.rooms[client.roomId] = make(map[string]*Client)
				}
				m.rooms[client.roomId][client.clientID] = client
				roomClientCount = len(m.rooms[client.roomId])
			}
			totalClients := len(m.clients)
			m.mu.Unlock()
			m.logger.Info().
				Str("client_id", client.clientID).
				Str("room_id", client.roomId).
				Int("total_clients", totalClients).
				Int("room_clients", roomClientCount).
				Msg("Client connected")

		case client := <-m.unregister:
			m.mu.Lock()
			var roomID string
			var userID int
			var nickName string
			if _, ok := m.clients[client.clientID]; ok {
				delete(m.clients, client.clientID)
				if client.roomId != "" && m.rooms[client.roomId] != nil {
					roomID = client.roomId
					userID = client.userId
					nickName = client.nickName
					delete(m.rooms[client.roomId], client.clientID)
					if len(m.rooms[client.roomId]) == 0 {
						delete(m.rooms, client.roomId)
					}
				}
			}
			totalClients := len(m.clients)
			m.mu.Unlock()

			// 广播 user_leave 给房间其他人
			if roomID != "" {
				payload := fmt.Sprintf(`{"room_id":"%s","user_id":%d,"nickname":"%s"}`, roomID, userID, nickName)
				msg, _ := json.Marshal(map[string]any{
					"action":  "user_leave",
					"payload": json.RawMessage(payload),
				})
				m.SendToRoom(roomID, msg)
			}

			m.logger.Info().
				Str("client_id", client.clientID).
				Str("room_id", roomID).
				Int("total_clients", totalClients).
				Msg("Client disconnected")

		case message := <-m.broadcast:
			m.mu.RLock()
			clientsCopy := make([]*Client, 0, len(m.clients))
			for _, client := range m.clients {
				clientsCopy = append(clientsCopy, client)
			}
			m.mu.RUnlock()

			for _, client := range clientsCopy {
				client.Send(message)
			}
		}
	}
}

func (m *Manager) Broadcast(message []byte) {
	m.broadcast <- message
}

func (m *Manager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

func (m *Manager) GetRoomClientCount(roomId string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if room, ok := m.rooms[roomId]; ok {
		return len(room)
	}
	return 0
}

func (m *Manager) Register(client *Client) {
	m.register <- client
}

func (m *Manager) SetRoomToken(roomId string, newToken string, expireAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.roomsToken[roomId]; ok {
		existing.OldToken = existing.NewToken
		existing.NewToken = newToken
		existing.ExpireAt = expireAt
	} else {
		m.roomsToken[roomId] = &RoomToken{
			OldToken: "",
			NewToken: newToken,
			ExpireAt: expireAt,
		}
	}
}

func (m *Manager) GetRoomToken(roomId string) (*RoomToken, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.roomsToken[roomId]
	return token, ok
}

func (m *Manager) RemoveRoomToken(roomId string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roomsToken, roomId)
}

func (m *Manager) GetUserNickname(userId int) string {
	if userId <= 0 || m.userRepo == nil {
		return ""
	}

	ctx := context.Background()
	user, err := m.userRepo.GetByID(ctx, userId)
	if err != nil {
		m.logger.Error().Err(err).Int("user_id", userId).Msg("Failed to get user nickname")
		return ""
	}

	return user.Nickname
}

func (m *Manager) SendToRoom(roomID string, message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return
	}

	for _, client := range room {
		client.Send(message)
	}
}

func (m *Manager) SendToRoomExcept(roomID string, excludeClientID string, message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return
	}

	for clientID, client := range room {
		if clientID == excludeClientID {
			continue
		}
		client.Send(message)
	}
}

type UserInfo struct {
	UserID   int    `json:"user_id"`
	NickName string `json:"nickname"`
}

func (m *Manager) GetRoomUsers(roomID string) []UserInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil
	}

	users := make([]UserInfo, 0, len(room))
	for _, client := range room {
		nickName := client.nickName
		users = append(users, UserInfo{
			UserID:   client.userId,
			NickName: nickName,
		})
	}

	return users
}

func (m *Manager) loadRoomsFromDB() {
	if m.chatRoomRepo == nil {
		m.logger.Warn().Msg("ChatRoomRepository not set, skipping room loading")
		return
	}

	ctx := context.Background()
	rooms, _, err := m.chatRoomRepo.GetAll(ctx, 1, 1000)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to load rooms from database")
		return
	}

	m.mu.Lock()
	for _, room := range rooms {
		roomID := fmt.Sprintf("%d", room.ID)
		if m.rooms[roomID] == nil {
			m.rooms[roomID] = make(map[string]*Client)
		}
		// 初始化空的 roomsToken 或者从其他地方加载 token
		if _, exists := m.roomsToken[roomID]; !exists {
			m.roomsToken[roomID] = &RoomToken{
				OldToken: "",
				NewToken: m.generateToken(),
				ExpireAt: time.Now().Add(5 * time.Minute),
			}
		}
	}
	m.mu.Unlock()

	m.logger.Info().Int("room_count", len(rooms)).Msg("Rooms loaded from database successfully")
}

func (m *Manager) startTokenRefreshTimer() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.refreshRoomTokens()
	}
}

func (m *Manager) refreshRoomTokens() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for roomID, token := range m.roomsToken {
		token.OldToken = token.NewToken
		token.NewToken = m.generateToken()
		token.ExpireAt = time.Now().Add(5 * time.Minute)
		m.logger.Debug().Str("room_id", roomID).Str("new_token", token.NewToken).Msg("Room token refreshed")
	}

	m.logger.Info().Int("token_count", len(m.roomsToken)).Msg("All room tokens refreshed")
}

func (m *Manager) generateToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (m *Manager) JoinRoom(clientID string, roomID string, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[clientID]
	if !ok {
		return fmt.Errorf("client not found")
	}

	// 校验 token
	rt, ok := m.roomsToken[roomID]
	if !ok {
		return fmt.Errorf("room not found")
	}
	if token != rt.NewToken && token != rt.OldToken {
		return fmt.Errorf("invalid token")
	}

	// 从旧房间移除
	if client.roomId != "" && m.rooms[client.roomId] != nil {
		delete(m.rooms[client.roomId], client.clientID)
		if len(m.rooms[client.roomId]) == 0 {
			delete(m.rooms, client.roomId)
		}
	}

	// 加入新房间
	if m.rooms[roomID] == nil {
		m.rooms[roomID] = make(map[string]*Client)
	}
	m.rooms[roomID][client.clientID] = client
	client.roomId = roomID

	return nil
}

func (m *Manager) LeaveRoom(clientID string, roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[clientID]
	if !ok {
		return fmt.Errorf("client not found")
	}

	if client.roomId != roomID {
		return fmt.Errorf("not in this room")
	}

	if m.rooms[roomID] != nil {
		delete(m.rooms[roomID], client.clientID)
		if len(m.rooms[roomID]) == 0 {
			delete(m.rooms, roomID)
		}
	}

	client.roomId = ""
	return nil
}

func (m *Manager) SaveMessage(message *models.Message) error {
	if m.messageRepo == nil {
		return fmt.Errorf("message repository not initialized")
	}
	ctx := context.Background()
	return m.messageRepo.Create(ctx, message)
}
