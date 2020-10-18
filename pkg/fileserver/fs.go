package fileserver

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
)

//
type FileServer struct {
	PublicPath string
	StaticPath string

	CacheControl  *CacheControl
	CookieControl *CookieControl
}

// CacheControl holds description of standard http cache Header
type CacheControl struct {
	Cache string
}

// CookieControl holds description of cookie policy to use.
type CookieControl struct {
	Name      string
	ValueFunc func() string
	TTL       time.Duration
}

// Handler returns new chi handler to server static content from a
func (fsd *FileServer) Handler() func(router chi.Router) {
	return func(router chi.Router) {
		fs := http.FileServer(http.Dir(fsd.StaticPath))

		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("cache-control", "no-cache")

			if fsd.CookieControl != nil {
				cname := fsd.CookieControl.Name
				c, _ := r.Cookie(cname)

				if c == nil {
					http.SetCookie(w, &http.Cookie{
						Name:    fsd.CookieControl.Name,
						Value:   fsd.CookieControl.ValueFunc(),
						Expires: time.Now().Add(fsd.CookieControl.TTL),
					})
				}
			}

			if _, err := os.Stat(fsd.StaticPath + r.RequestURI); os.IsNotExist(err) {
				http.StripPrefix(fsd.PublicPath, fs).ServeHTTP(w, r)
			} else {
				fs.ServeHTTP(w, r)
			}
		})
	}
}
