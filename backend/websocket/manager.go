package websocket

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	manager  *Manager
	send     chan []byte
	clientID string
}

type Manager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
	logger     zerolog.Logger
}

func NewManager(logger zerolog.Logger) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		logger:     logger,
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.clientID] = client
			m.mu.Unlock()
			m.logger.Info().Str("client_id", client.clientID).Int("total_clients", len(m.clients)).Msg("Client connected")

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.clientID]; ok {
				close(client.send)
				delete(m.clients, client.clientID)
			}
			m.mu.Unlock()
			m.logger.Info().Str("client_id", client.clientID).Int("total_clients", len(m.clients)).Msg("Client disconnected")

		case message := <-m.broadcast:
			m.mu.RLock()
			for _, client := range m.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(m.clients, client.clientID)
				}
			}
			m.mu.RUnlock()
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

func (m *Manager) Register(client *Client) {
	m.register <- client
}
