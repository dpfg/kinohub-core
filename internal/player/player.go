package player

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Player is a middleman between the websocket connection and the hub.
type Player struct {
	pid string

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// playList is a list of media entries to play sequentially
	playList *PList
}

type message struct {
	TypeID string      `json:"type_id,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Player) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Player) writePump() {
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
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

func (c *Player) sendPlay() {
	c.sendMessage(message{TypeID: "play"})
}

func (c *Player) sendPause() {
	c.sendMessage(message{TypeID: "pause"})
}

func (c *Player) sendStop() {
	c.sendMessage(message{TypeID: "stop"})
}

func (c *Player) sendSetSource(entry *MediaEntry) {
	c.sendMessage(message{TypeID: "set-source", Data: entry})
}

func (c *Player) sendRewind(duration int) {
	c.sendMessage(message{
		TypeID: "rewind",
		Data: struct {
			Duration int `json:"duration,omitempty"`
		}{
			Duration: duration,
		},
	})
}

func (c *Player) sendMessage(msg message) {
	b, _ := json.Marshal(msg)
	c.send <- b
}

func serveWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	pid := r.URL.Query().Get("pid")
	if pid == "" {
		return
	}

	hub.Disconnect(pid)

	player := &Player{
		pid:      pid,
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		playList: NewPlayList([]MediaEntry{}),
	}

	player.hub.register <- player

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go player.writePump()
	go player.readPump()
}
