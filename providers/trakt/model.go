package trakt

import "time"

type MyShow struct {
	FirstAired time.Time `json:"first_aired"`
	Episode    struct {
		Season int    `json:"season"`
		Number int    `json:"number"`
		Title  string `json:"title"`
		Ids    struct {
			Trakt  int         `json:"trakt"`
			Tvdb   int         `json:"tvdb"`
			Imdb   string      `json:"imdb"`
			Tmdb   int         `json:"tmdb"`
			Tvrage interface{} `json:"tvrage"`
		} `json:"ids"`
	} `json:"episode"`
	Show struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt  int    `json:"trakt"`
			Slug   string `json:"slug"`
			Tvdb   int    `json:"tvdb"`
			Imdb   string `json:"imdb"`
			Tmdb   int    `json:"tmdb"`
			Tvrage int    `json:"tvrage"`
		} `json:"ids"`
	} `json:"show"`
}
