package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func NewClient(conn *websocket.Conn, manager *Manager, userId int, nickName string) *Client {
	return &Client{
		conn:     conn,
		manager:  manager,
		send:     make(chan []byte, maxMessageSize),
		clientID: uuid.New().String(),

		// 业务逻辑
		isAuthenticated: false,
		roomId:          "",
		userId:          userId,
		nickName:        nickName,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.manager.logger.Error().Err(err).Str("client_id", c.clientID).Msg("WebSocket read error")
			}
			break
		}

		//message 解析
		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			c.manager.logger.Error().Err(err).Str("client_id", c.clientID).Msg("WebSocket message parse error")
			// continue
		}
		c.manager.logger.Info().Str("client_id", c.clientID).Msg("WebSocket message received: " + string(message))

		// 处理 ping 消息，返回 pong 响应
		if msg.Action == "ping" {
			pongMsg := Message{
				Action:  "pong",
				Payload: json.RawMessage(`{}`),
			}
			pongData, _ := json.Marshal(pongMsg)
			c.send <- pongData
			continue
		}

		c.manager.Broadcast(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}
