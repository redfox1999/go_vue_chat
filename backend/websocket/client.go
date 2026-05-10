package websocket

import (
	"backend/dto"
	"backend/models"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second // 写入超时时间
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

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

func errMsg(action, msg string) []byte {
	p, _ := json.Marshal(map[string]string{"error": msg})
	b, _ := json.Marshal(dto.Message{Action: action, Payload: p})
	return b
}

func NewClient(conn *websocket.Conn, manager *Manager, userId int, nickName string) *Client {
	return &Client{
		conn:     conn,
		manager:  manager,
		send:     make(chan []byte, 16),
		clientID: uuid.New().String(),

		// 业务逻辑
		roomId:   "",
		userId:   userId,
		nickName: nickName,
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.manager.unregister <- c
		close(c.send)
		c.conn.Close()
	})
}

func (c *Client) Send(message []byte) {
	select {
	case c.send <- message:
		// 正常发送
	default:
		// 如果 16 个位置都占满了，说明该客户端网络极差，直接踢掉
		c.manager.logger.Warn().Str("client_id", c.clientID).Msg("Client send buffer full, dropping")
		c.Close()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.Close()
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
		var msg dto.Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			c.manager.logger.Error().Err(err).Str("client_id", c.clientID).Msg("WebSocket message parse error")
			// continue
		}
		c.manager.logger.Info().Str("client_id", c.clientID).Msg("WebSocket message received: " + string(message))

		switch msg.Action {
		case "ping":
			pongMsg := dto.Message{
				Action:  "pong",
				Payload: json.RawMessage(`{}`),
			}
			pongData, _ := json.Marshal(pongMsg)
			c.Send(pongData)

		case "join":
			var payload dto.JoinPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				errMsg, _ := json.Marshal(dto.Message{
					Action:  "join_error",
					Payload: json.RawMessage(`{"error":"invalid payload"}`),
				})
				c.Send(errMsg)
				break
			}
			if err := c.manager.JoinRoom(c.clientID, payload.RoomID, payload.Token); err != nil {
				errPayload, _ := json.Marshal(map[string]string{"error": err.Error()})
				errMsg, _ := json.Marshal(dto.Message{
					Action:  "join_error",
					Payload: errPayload,
				})
				c.Send(errMsg)
				break
			}
			joinPayload := fmt.Sprintf(`{"room_id":"%s","user_id":%d,"nickname":"%s"}`, payload.RoomID, c.userId, c.nickName)
			joinBroadcast, _ := json.Marshal(dto.Message{
				Action:  "user_join",
				Payload: json.RawMessage(joinPayload),
			})
			c.manager.SendToRoomExcept(payload.RoomID, c.clientID, joinBroadcast)

			okPayload, _ := json.Marshal(map[string]string{"room_id": payload.RoomID})
			okMsg, _ := json.Marshal(dto.Message{
				Action:  "join_ok",
				Payload: okPayload,
			})
			c.Send(okMsg)

		case "leave":
			var payload dto.LeavePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				errMsg, _ := json.Marshal(dto.Message{
					Action:  "leave_error",
					Payload: json.RawMessage(`{"error":"invalid payload"}`),
				})
				c.Send(errMsg)
				break
			}
			leavePayload := fmt.Sprintf(`{"room_id":"%s","user_id":%d,"nickname":"%s"}`, payload.RoomID, c.userId, c.nickName)
			leaveBroadcast, _ := json.Marshal(dto.Message{
				Action:  "user_leave",
				Payload: json.RawMessage(leavePayload),
			})
			c.manager.SendToRoom(payload.RoomID, leaveBroadcast)

			if err := c.manager.LeaveRoom(c.clientID, payload.RoomID); err != nil {
				errPayload, _ := json.Marshal(map[string]string{"error": err.Error()})
				errMsg, _ := json.Marshal(dto.Message{
					Action:  "leave_error",
					Payload: errPayload,
				})
				c.Send(errMsg)
				break
			}
			okMsg, _ := json.Marshal(dto.Message{
				Action:  "leave_ok",
				Payload: json.RawMessage(`{}`),
			})
			c.Send(okMsg)

		case "chat":
			var payload dto.ChatPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				c.Send(errMsg("chat_error", "invalid payload"))
				break
			}
			if c.roomId != payload.RoomID {
				c.Send(errMsg("chat_error", "not in this room"))
				break
			}

			roomID, _ := strconv.Atoi(payload.RoomID)
			message := &models.Message{
				RoomID:   roomID,
				Sender:   c.userId,
				Nickname: c.nickName,
				Notify:   "",
				Message:  payload.Content,
				SendTime: time.Now(),
			}
			if err := c.manager.SaveMessage(message); err != nil {
				c.manager.logger.Error().Err(err).Msg("Failed to save message to database")
			}

			chatPayload, _ := json.Marshal(map[string]any{
				"room_id":  payload.RoomID,
				"user_id":  c.userId,
				"nickname": c.nickName,
				"content":  payload.Content,
			})
			chatBroadcast, _ := json.Marshal(dto.Message{
				Action:  "chat",
				Payload: chatPayload,
			})
			c.manager.SendToRoomExcept(payload.RoomID, c.clientID, chatBroadcast)

		default:
			c.manager.Broadcast(message)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)

			if err != nil {
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
