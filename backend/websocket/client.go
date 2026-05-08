package websocket

import (
	"encoding/json"
	"fmt"
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

type JoinPayload struct {
	RoomID string `json:"room_id"`
	Token  string `json:"token"`
}

type LeavePayload struct {
	RoomID string `json:"room_id"`
}

type ChatPayload struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

func errMsg(action, msg string) []byte {
	p, _ := json.Marshal(map[string]string{"error": msg})
	b, _ := json.Marshal(Message{Action: action, Payload: p})
	return b
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

		// 处理 join 消息，加入指定聊天室
		if msg.Action == "join" {
			var payload JoinPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				errMsg, _ := json.Marshal(Message{
					Action:  "join_error",
					Payload: json.RawMessage(`{"error":"invalid payload"}`),
				})
				c.send <- errMsg
				continue
			}
			if err := c.manager.JoinRoom(c.clientID, payload.RoomID, payload.Token); err != nil {
				errPayload, _ := json.Marshal(map[string]string{"error": err.Error()})
				errMsg, _ := json.Marshal(Message{
					Action:  "join_error",
					Payload: errPayload,
				})
				c.send <- errMsg
				continue
			}
			// 广播 user_join 给房间所有人
			joinPayload := fmt.Sprintf(`{"room_id":"%s","user_id":%d,"nickname":"%s"}`, payload.RoomID, c.userId, c.nickName)
			joinBroadcast, _ := json.Marshal(Message{
				Action:  "user_join",
				Payload: json.RawMessage(joinPayload),
			})
			c.manager.SendToRoom(payload.RoomID, joinBroadcast)

			okPayload, _ := json.Marshal(map[string]string{"room_id": payload.RoomID})
			okMsg, _ := json.Marshal(Message{
				Action:  "join_ok",
				Payload: okPayload,
			})
			c.send <- okMsg
			continue
		}

		// 处理 leave 消息，离开当前聊天室
		if msg.Action == "leave" {
			var payload LeavePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				errMsg, _ := json.Marshal(Message{
					Action:  "leave_error",
					Payload: json.RawMessage(`{"error":"invalid payload"}`),
				})
				c.send <- errMsg
				continue
			}
			// 广播 user_leave 给房间所有人（在离开之前，确保能拿到用户信息）
			leavePayload := fmt.Sprintf(`{"room_id":"%s","user_id":%d,"nickname":"%s"}`, payload.RoomID, c.userId, c.nickName)
			leaveBroadcast, _ := json.Marshal(Message{
				Action:  "user_leave",
				Payload: json.RawMessage(leavePayload),
			})
			c.manager.SendToRoom(payload.RoomID, leaveBroadcast)

			if err := c.manager.LeaveRoom(c.clientID, payload.RoomID); err != nil {
				errPayload, _ := json.Marshal(map[string]string{"error": err.Error()})
				errMsg, _ := json.Marshal(Message{
					Action:  "leave_error",
					Payload: errPayload,
				})
				c.send <- errMsg
				continue
			}
			okMsg, _ := json.Marshal(Message{
				Action:  "leave_ok",
				Payload: json.RawMessage(`{}`),
			})
			c.send <- okMsg
			continue
		}

		// 处理 chat 消息，房间内群发
		if msg.Action == "chat" {
			var payload ChatPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				c.send <- errMsg("chat_error", "invalid payload")
				continue
			}
			if c.roomId != payload.RoomID {
				c.send <- errMsg("chat_error", "not in this room")
				continue
			}
			chatPayload, _ := json.Marshal(map[string]any{
				"room_id":  payload.RoomID,
				"user_id":  c.userId,
				"nickname": c.nickName,
				"content":  payload.Content,
			})
			chatBroadcast, _ := json.Marshal(Message{
				Action:  "chat",
				Payload: chatPayload,
			})
			c.manager.SendToRoom(payload.RoomID, chatBroadcast)
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
