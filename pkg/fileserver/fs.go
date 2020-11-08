package fileserver

import (
	"github.com/markbates/pkger"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// A EmbeddedFolder implements http.FileSystem using the pkger and embedded static resources.
type EmbeddedFolder struct {
	Root string
}

// Open implements http.FileSystem using pkger.Open
func (folder EmbeddedFolder) Open(name string) (http.File, error) {
	// pkger should always be used with unix-style path separator
	path := strings.ReplaceAll(filepath.Join(folder.Root, name), string(os.PathSeparator), "/")
	f, err := pkger.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s != nil && s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err = pkger.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

// Handler returns new chi handler to server static content from a
func (fsd *FileServer) Handler() func(router chi.Router) {
	return func(router chi.Router) {
		fs := http.FileServer(EmbeddedFolder{fsd.StaticPath})

		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {

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

			if fsd.CacheControl != nil {
				w.Header().Add("Cache-Control", fsd.CacheControl.Cache)
			}

			http.StripPrefix(fsd.PublicPath, fs).ServeHTTP(w, r)
		})
	}
}
