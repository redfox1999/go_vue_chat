package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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

	isAuthenticated bool
	roomId          string
	userId          int
	nickName        string
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
}

func NewManager(logger zerolog.Logger, userRepo repository.UserRepository, chatRoomRepo repository.ChatRoomRepository) *Manager {
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
			var totalClients int
			if _, ok := m.clients[client.clientID]; ok {
				close(client.send)
				delete(m.clients, client.clientID)
				if client.roomId != "" && m.rooms[client.roomId] != nil {
					delete(m.rooms[client.roomId], client.clientID)
					if len(m.rooms[client.roomId]) == 0 {
						delete(m.rooms, client.roomId)
					}
				}
			}
			totalClients = len(m.clients)
			m.mu.Unlock()
			m.logger.Info().
				Str("client_id", client.clientID).
				Str("room_id", client.roomId).
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
				select {
				case client.send <- message:
				default:
					m.mu.Lock()
					close(client.send)
					delete(m.clients, client.clientID)
					if client.roomId != "" && m.rooms[client.roomId] != nil {
						delete(m.rooms[client.roomId], client.clientID)
						if len(m.rooms[client.roomId]) == 0 {
							delete(m.rooms, client.roomId)
						}
					}
					m.mu.Unlock()
				}
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
