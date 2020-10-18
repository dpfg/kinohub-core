package player

import "github.com/sirupsen/logrus"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered players.
	players map[*Player]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Player

	// Unregister requests from clients.
	unregister chan *Player
	logger     *logrus.Entry
}

func newHub(logger *logrus.Entry) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		players:    make(map[*Player]bool),
		logger:     logger,
	}
}

func (h *Hub) run() {
	for {
		select {
		case player := <-h.register:
			h.players[player] = true
		case player := <-h.unregister:
			if _, ok := h.players[player]; ok {
				delete(h.players, player)
				close(player.send)
			}
		case message := <-h.broadcast:
			for client := range h.players {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.players, client)
				}
			}
		}
	}
}

func (h *Hub) Disconnect(pid string) bool {
	for p := range h.players {
		if p.pid == pid {
			h.logger.Debugf("Disconnecting player: %s\n", pid)

			delete(h.players, p)
			close(p.send)
			return true
		}
	}
	return false
}
