package services

import (
	"net/http"

	"github.com/go-chi/chi"
)

type SystemModule struct {
}

func (mod SystemModule) Handler() http.Handler {
	router := chi.NewRouter()

	router.Delete("/cache", func(w http.ResponseWriter, req *http.Request) {

	})

	return router
}
