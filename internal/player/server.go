package player

import (
	"net/http"
	"strconv"

	"github.com/dpfg/kinohub-core/pkg/fileserver"
	"github.com/dpfg/kinohub-core/util"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Server is an entry entity to provide player functionality based on JS-player with ability to control playback though WebSocket
type Server struct {
	hub *Hub
}

func NewServer() *Server {
	hub := newHub()
	go hub.run()
	return &Server{hub: hub}
}

// Handler returns chi.Router registrar to handle player-related http endpoints
func (srv Server) Handler() func(r chi.Router) {

	return func(router chi.Router) {

		router.HandleFunc("/ui/pws/", func(w http.ResponseWriter, r *http.Request) {
			serveWebSocket(srv.hub, w, r)
		})

		router.Route("/ui/player", fileserver.FileServer("/ui/player/", "./web/player"))

		router.Route("/api/players/", func(r chi.Router) {
			r.Get("/", srv.httpListAll)

			r.Route("/{pid}", func(r chi.Router) {
				r.Post("/pause", srv.httpPause)
				r.Post("/play", srv.httpPlay)
				r.Get("/plist", srv.httpPlayList)
				r.Post("/plist", srv.httpPlayListAdd)
				r.Post("/plist/commands/select", srv.httpPlayListSelect)
			})

		})
	}
}

func (srv Server) findPlayer(pid string) *Player {
	for p := range srv.hub.players {
		if p.pid == pid {
			return p
		}
	}

	return nil
}

func (srv Server) httpListAll(w http.ResponseWriter, r *http.Request) {
	list := make([]string, 0)
	for player := range srv.hub.players {
		list = append(list, player.pid)
	}
	render.JSON(w, r, list)
}

func (srv Server) httpPlay(w http.ResponseWriter, r *http.Request) {
	p := srv.findPlayer(chi.URLParam(r, "pid"))
	p.sendPlay()
	render.Status(r, http.StatusAccepted)
}

func (srv Server) httpPause(w http.ResponseWriter, r *http.Request) {
	p := srv.findPlayer(chi.URLParam(r, "pid"))
	p.sendPause()
	render.Status(r, http.StatusAccepted)
}

func (srv Server) httpPlayList(w http.ResponseWriter, r *http.Request) {
	p := srv.findPlayer(chi.URLParam(r, "pid"))
	render.JSON(w, r, p.playList)
}

func (srv Server) httpPlayListAdd(w http.ResponseWriter, r *http.Request) {

	media := MediaEntry{}
	err := render.DecodeJSON(r.Body, &media)
	if err != nil {
		util.BadRequest(w, r, err)
		return
	}

	player := srv.findPlayer(chi.URLParam(r, "pid"))
	index := player.playList.AddEntry(media)

	sel, _ := strconv.ParseBool(r.URL.Query().Get("select"))
	if sel {
		entry := player.playList.Select(index)
		player.sendSetSource(entry)
	}

	render.JSON(w, r, player.playList)
}

func (srv Server) httpPlayListSelect(w http.ResponseWriter, r *http.Request) {
	body := &struct {
		Position int `json:"position,omitempty"`
	}{}

	err := render.DecodeJSON(r.Body, body)
	if err != nil {
		util.BadRequest(w, r, err)
		return
	}

	p := srv.findPlayer(chi.URLParam(r, "pid"))

	entry := p.playList.Select(body.Position)
	p.sendSetSource(entry)

	render.Status(r, http.StatusAccepted)
}
