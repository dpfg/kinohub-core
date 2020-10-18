package util

import (
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

func JoinURL(parts ...interface{}) string {
	url := ""
	for _, p := range parts {
		if _, ok := p.(string); ok {
			url = append(url, p.(string))
		}
		if _, ok := p.(int); ok {
			url = append(url, strconv.Itoa(p.(int)))
		}
	}
	return url
}

func append(p1, p2 string) string {
	if p1 == "" {
		return p1 + p2
	}
	return p1 + "/" + p2
}

func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)
	renderError(w, r, err)
}

func NotFound(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusNotFound)
	renderError(w, r, err)
}

func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusBadRequest)
	renderError(w, r, err)
}

func BadGateway(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusBadGateway)
	renderError(w, r, err)
}

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	render.JSON(w, r, struct {
		Msg string `json:"message"`
	}{Msg: err.Error()})
}
