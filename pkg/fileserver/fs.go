package fileserver

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

//
type FileServer struct {
	PublicPath string
	StaticPath string

	CacheControl CacheControl
}

// CacheControl holds description of standard http cache Header
type CacheControl struct {
	Cache string
}

// Handler returns new chi handler to server static content from a
func (fsd *FileServer) Handler() func(router chi.Router) {
	return func(router chi.Router) {
		fs := http.FileServer(http.Dir(fsd.StaticPath))

		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("cache-control", "no-cache")

			if _, err := os.Stat(fsd.StaticPath + r.RequestURI); os.IsNotExist(err) {
				http.StripPrefix(fsd.PublicPath, fs).ServeHTTP(w, r)
			} else {
				fs.ServeHTTP(w, r)
			}
		})
	}
}
