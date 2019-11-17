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

func PadLeft(str, pad string, lenght int) string {
	for {
		if len(str) >= lenght {
			return str
		}
		str = pad + str
	}
}

func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, struct {
		Msg string `json:"message"`
	}{Msg: err.Error()})
}
