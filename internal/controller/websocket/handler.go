package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/geolav/QEasyApp/internal/port"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // в продакшене тут проверяем домен
	},
}

type WSHandler struct {
	hub       *Hub
	sessionUC port.SessionUseCase
	quizUC    port.QuizUseCase
}

func NewHandler(hub *Hub, sessionUC port.SessionUseCase, quizUC port.QuizUseCase) *WSHandler {
	return &WSHandler{hub: hub, sessionUC: sessionUC, quizUC: quizUC}
}

// Connect — точка входа, апгрейдим HTTP → WebSocket
func (h *WSHandler) Connect(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.GetString("user_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &Client{
		SessionID: sessionID,
		UserID:    userID,
		Send:      make(chan []byte, 256),
		hub:       h.hub,
		conn:      conn,
	}

	h.hub.Register <- client

	// запускаем чтение и запись в отдельных горутинах
	go h.writePump(client, conn)
	go h.readPump(client, conn, h)
}

// writePump — отправляет сообщения клиенту из канала Send
func (h *WSHandler) writePump(client *Client, conn *websocket.Conn) {
	defer conn.Close()
	for msg := range client.Send {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// readPump — читает сообщения от клиента
func (h *WSHandler) readPump(client *Client, conn *websocket.Conn, handler *WSHandler) {
	defer func() {
		h.hub.Unregister <- client
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// парсим входящее сообщение
		var event struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		handler.handleEvent(client, event.Type, event.Payload)
	}
}

// handleEvent — роутер событий
func (h *WSHandler) handleEvent(client *Client, eventType string, payload json.RawMessage) {
	switch eventType {
	case "start_session":
		h.handleStartSession(client)
	case "next_question":
		h.handleNextQuestion(client, payload)
	case "submit_answer":
		h.handleSubmitAnswer(client, payload)
	case "finish_session":
		h.handleFinishSession(client)
	}
}
