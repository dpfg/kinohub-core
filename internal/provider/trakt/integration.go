package trakt

import (
	"context"
	"net/http"

	httpu "github.com/dpfg/kinohub-core/pkg/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Integration with Trakt.tv
type Integration struct {
	Client *Client
}

// Handler with defined routes for Trakt integration
func (trakt *Integration) Handler() http.Handler {
	router := chi.NewRouter()

	router.Get("/trending", func(w http.ResponseWriter, req *http.Request) {
		shows, err := trakt.Client.TrendingShows()
		if err != nil {
			httpu.InternalError(w, req, err)
			return
		}

		render.JSON(w, req, shows)
	})

	router.Get("/signin", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, trakt.Client.AuthCodeURL(), http.StatusTemporaryRedirect)
	})

	router.Get("/exchange", func(w http.ResponseWriter, req *http.Request) {
		_, err := trakt.Client.Exchange(context.Background(), req.URL.Query().Get("code"))
		if err != nil {
			httpu.InternalError(w, req, err)
			return
		}

		http.Redirect(w, req, "/status", http.StatusTemporaryRedirect)
	})

	router.Get("/status", func(w http.ResponseWriter, req *http.Request) {
		s, err := trakt.Client.Settings()
		if err != nil {
			httpu.InternalError(w, req, err)
			return
		}
		render.JSON(w, req, s)
	})

	return router
}
