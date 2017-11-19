package domain

import (
	"time"
)

type Movie struct {
	TmdbID   int    `json:"tmdb_id,omitempty"`
	Title    string `json:"title,omitempty"`
	Year     int    `json:"year,omitempty"`
	Overview string `json:"overview,omitempty"`
}

type Show struct {
	TmdbID   int      `json:"tmdb_id,omitempty"`
	Title    string   `json:"title,omitempty"`
	Year     int      `json:"year,omitempty"`
	Overview string   `json:"overview,omitempty"`
	Seasons  []Season `json:"seasons,omitempty"`
}

type Season struct {
	TmdbID     int       `json:"tmdb_id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Number     int       `json:"number,omitempty"`
	AirDate    string    `json:"air_date"`
	Episodes   []Episode `json:"episodes,omitempty"`
	PosterPath string    `json:"poster_path,omitempty"`
}

type Episode struct {
	TmdbID     int       `json:"tmdb_id,omitempty"`
	Season     int       `json:"season,omitempty"`
	Number     int       `json:"number,omitempty"`
	Title      string    `json:"title,omitempty"`
	Overview   string    `json:"overview,omitempty"`
	FirstAired time.Time `json:"first_aired,omitempty"`
	Files      []File    `json:"files,omitempty"`
}

type File struct {
	Quality string `json:"quality"`
	URL     struct {
		HTTP string `json:"http"`
		Hls  string `json:"hls"`
		Hls4 string `json:"hls4"`
	} `json:"url"`
}

const (
	TypeSerial  = "SERIAL"
	TypeMovie   = "MOVIE"
	TypeUnknown = "UNKNOWN"
)

type SearchResult struct {
	Type       string `json:"type,omitempty"`
	TmdbID     int    `json:"tmdb_id,omitempty"`
	Title      string `json:"title,omitempty"`
	PosterPath string `json:"poster_path,omitempty"`
}
