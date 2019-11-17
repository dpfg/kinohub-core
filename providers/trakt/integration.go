package trakt

import (
	"context"
	"net/http"

	"github.com/dpfg/kinohub-core/util"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type Integration struct {
	Client *Client
}

func (trakt *Integration) Handler() http.Handler {
	router := chi.NewRouter()

	router.Get("/trending", func(w http.ResponseWriter, req *http.Request) {
		shows, err := trakt.Client.GetTrendingShows()
		if err != nil {
			util.InternalError(w, req, err)
			return
		}

		render.JSON(w, req, shows)
	})

	router.Get("/signin", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, trakt.Client.GetAuthCodeURL(), http.StatusTemporaryRedirect)
	})

	router.Get("/exchange", func(w http.ResponseWriter, req *http.Request) {
		_, err := trakt.Client.Exchange(context.Background(), req.URL.Query().Get("code"))
		if err != nil {
			util.InternalError(w, req, err)
			return
		}

		http.Redirect(w, req, "/status", http.StatusTemporaryRedirect)
	})

	router.Get("/status", func(w http.ResponseWriter, req *http.Request) {

		render.PlainText(w, req, "")
	})

	return router
}
