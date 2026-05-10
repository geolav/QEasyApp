package websocket

import "sync"

// Client — одно WebSocket подключение
type Client struct {
	SessionID string
	UserID    string
	Send      chan []byte
	hub       *Hub
	conn      interface{ WriteMessage(int, []byte) error }
}

// Message — сообщение которое гуляет между клиентами
type Message struct {
	SessionID string
	Data      []byte
}

// Hub — хранит все подключения и раздаёт сообщения
type Hub struct {
	mutex      sync.RWMutex
	sessions   map[string]map[*Client]bool // sessionID → список клиентов
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		sessions:   make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

// Run — главный цикл хаба, запускается в отдельной горутине
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			if h.sessions[client.SessionID] == nil {
				h.sessions[client.SessionID] = make(map[*Client]bool)
			}
			h.sessions[client.SessionID][client] = true
			h.mutex.Unlock()

		case client := <-h.Unregister:
			h.mutex.Lock()
			if clients, ok := h.sessions[client.SessionID]; ok {
				delete(clients, client)
				close(client.Send)
				if len(clients) == 0 {
					delete(h.sessions, client.SessionID)
				}
			}
			h.mutex.Unlock()

		case msg := <-h.Broadcast:
			h.mutex.RLock()
			clients := h.sessions[msg.SessionID]
			for client := range clients {
				select {
				case client.Send <- msg.Data:
				default:
					// канал заполнен — клиент завис, отключаем
					close(client.Send)
					delete(clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}
