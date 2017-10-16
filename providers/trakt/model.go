package trakt

import "time"

type MyShow struct {
	FirstAired time.Time `json:"first_aired,omitempty"`
	Episode    Episode   `json:"episode,omitempty"`
	Show       struct {
		Title string  `json:"title,omitempty"`
		Year  int     `json:"year,omitempty"`
		Ids   ShowIds `json:"ids,omitempty"`
	} `json:"show"`
}

type Episode struct {
	Season int        `json:"season,omitempty"`
	Number int        `json:"number,omitempty"`
	Title  string     `json:"title,omitempty"`
	Ids    EpisodeIds `json:"ids,omitempty"`
}

type ShowIds struct {
	Trakt  int    `json:"trakt,omitempty"`
	Slug   string `json:"slug,omitempty"`
	Tvdb   int    `json:"tvdb,omitempty"`
	Imdb   string `json:"imdb,omitempty"`
	Tmdb   int    `json:"tmdb,omitempty"`
	Tvrage int    `json:"tvrage,omitempty"`
}

type EpisodeIds struct {
	Trakt  int         `json:"trakt,omitempty"`
	Tvdb   int         `json:"tvdb,omitempty"`
	Imdb   string      `json:"imdb,omitempty"`
	Tmdb   int         `json:"tmdb,omitempty"`
	Tvrage interface{} `json:"tvrage,omitempty"`
}
