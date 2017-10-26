package domain

import "time"

type Show struct {
	TmdbID   int      `json:"tmdb_id,omitempty"`
	Title    string   `json:"title,omitempty"`
	Year     int      `json:"year,omitempty"`
	Overview string   `json:"overview,omitempty"`
	Seasons  []Season `json:"seasons,omitempty"`
}

type Season struct {
	TmdbID        int       `json:"tmdb_id,omitempty"`
	Number        int       `json:"number,omitempty"`
	EpisodeCount  int       `json:"episode_count,omitempty"`
	AiredEpisodes int       `json:"aired_episodes,omitempty"`
	Episodes      []Episode `json:"episodes,omitempty"`
}

type Episode struct {
	TmdbID     int       `json:"tmdb_id,omitempty"`
	Season     int       `json:"season,omitempty"`
	Number     int       `json:"number,omitempty"`
	Title      string    `json:"title,omitempty"`
	Overview   string    `json:"overview,omitempty"`
	FirstAired time.Time `json:"first_aired,omitempty"`
}
