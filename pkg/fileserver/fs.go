package fileserver

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// FileServer returns new chi handler to server static content from a
func FileServer(public, static string) func(router chi.Router) {
	return func(router chi.Router) {
		fs := http.FileServer(http.Dir(static))

		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if _, err := os.Stat(static + r.RequestURI); os.IsNotExist(err) {
				http.StripPrefix(public, fs).ServeHTTP(w, r)
			} else {
				fs.ServeHTTP(w, r)
			}
		})
	}
}
