package player

import (
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dpfg/kinohub-core/pkg/fileserver"
	httpu "github.com/dpfg/kinohub-core/pkg/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	UIDCookieName = "puid"
)

// Server is an entry entity to provide player functionality based on JS-player with ability to control playback though WebSocket
type Server struct {
	hub *Hub
}

func NewServer(logger *logrus.Entry) *Server {
	hub := newHub(logger)
	go hub.run()
	return &Server{hub: hub}
}

// Handler returns chi.Router registrar to handle player-related http endpoints
func (srv Server) Handler() func(r chi.Router) {

	return func(router chi.Router) {

		router.HandleFunc("/ui/pws/", func(w http.ResponseWriter, r *http.Request) {
			serveWebSocket(srv.hub, w, r)
		})

		fs := fileserver.FileServer{
			PublicPath: "/ui/player/",
			StaticPath: "/web/player",
			CacheControl: &fileserver.CacheControl{
				Cache: "no-cache",
			},
			CookieControl: &fileserver.CookieControl{
				Name: UIDCookieName,
				TTL:  time.Hour * 24 * 31 * 12,
				ValueFunc: func() string {
					uuid, err := uuid.NewUUID()
					if err != nil {
						return string(rand.Int31())
					}
					return uuid.String()
				},
			},
		}

		router.Route("/ui/player", fs.Handler())

		router.Route("/api/players/", func(r chi.Router) {
			r.Get("/", srv.httpListAll)

			r.Route("/{pid}", func(r chi.Router) {
				r.Post("/pause", srv.httpPause)
				r.Post("/play", srv.httpPlay)
				r.Post("/stop", srv.httpStop)
				r.Post("/rewind", srv.httpRewind)
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
	if p == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	p.sendPlay()
	render.Status(r, http.StatusAccepted)
}

func (srv Server) httpPause(w http.ResponseWriter, r *http.Request) {
	p := srv.findPlayer(chi.URLParam(r, "pid"))
	if p == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	p.sendPause()
	render.Status(r, http.StatusAccepted)
}

func (srv Server) httpStop(w http.ResponseWriter, r *http.Request) {
	p := srv.findPlayer(chi.URLParam(r, "pid"))
	if p == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	p.sendStop()
	render.Status(r, http.StatusAccepted)
}

func (srv Server) httpPlayList(w http.ResponseWriter, r *http.Request) {
	p := srv.findPlayer(chi.URLParam(r, "pid"))
	if p == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	render.JSON(w, r, p.playList)
}

func (srv Server) httpPlayListAdd(w http.ResponseWriter, r *http.Request) {

	media := MediaEntry{}
	err := render.DecodeJSON(r.Body, &media)
	if err != nil {
		httpu.BadRequest(w, r, err)
		return
	}

	player := srv.findPlayer(chi.URLParam(r, "pid"))
	if player == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	index := player.playList.AddEntry(media)

	sel, _ := strconv.ParseBool(r.URL.Query().Get("select"))
	if sel {
		entry := player.playList.Select(index)
		player.sendSetSource(entry)
		player.sendPlay()
	}

	render.JSON(w, r, player.playList)
}

func (srv Server) httpPlayListSelect(w http.ResponseWriter, r *http.Request) {
	body := &struct {
		Position int `json:"position,omitempty"`
	}{}

	err := render.DecodeJSON(r.Body, body)
	if err != nil {
		httpu.BadRequest(w, r, err)
		return
	}

	p := srv.findPlayer(chi.URLParam(r, "pid"))
	if p == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	entry := p.playList.Select(body.Position)
	p.sendSetSource(entry)

	render.Status(r, http.StatusAccepted)
}

func (srv Server) httpRewind(w http.ResponseWriter, r *http.Request) {
	body := &struct {
		Duration int `json:"duration,omitempty"`
	}{}

	err := render.DecodeJSON(r.Body, body)
	if err != nil {
		httpu.BadRequest(w, r, err)
		return
	}

	p := srv.findPlayer(chi.URLParam(r, "pid"))
	if p == nil {
		httpu.NotFound(w, r, errors.New("cannot find printer"))
		return
	}

	p.sendRewind(body.Duration)

	render.Status(r, http.StatusAccepted)
}
