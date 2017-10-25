package tmdb

import "encoding/json"
import "github.com/dpfg/kinohub-core/providers"

type TVShow struct {
	BackdropPath string `json:"backdrop_path"`
	CreatedBy    []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Gender      int    `json:"gender"`
		ProfilePath string `json:"profile_path"`
	} `json:"created_by"`
	EpisodeRunTime []int  `json:"episode_run_time"`
	FirstAirDate   string `json:"first_air_date"`
	Genres         []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage     string   `json:"homepage"`
	ID           int      `json:"id"`
	InProduction bool     `json:"in_production"`
	Languages    []string `json:"languages"`
	LastAirDate  string   `json:"last_air_date"`
	Name         string   `json:"name"`
	Networks     []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"networks"`
	NumberOfEpisodes    int      `json:"number_of_episodes"`
	NumberOfSeasons     int      `json:"number_of_seasons"`
	OriginCountry       []string `json:"origin_country"`
	OriginalLanguage    string   `json:"original_language"`
	OriginalName        string   `json:"original_name"`
	Overview            string   `json:"overview"`
	Popularity          float64  `json:"popularity"`
	PosterPath          string   `json:"poster_path"`
	ProductionCompanies []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	} `json:"production_companies"`
	Seasons []struct {
		AirDate      string `json:"air_date"`
		EpisodeCount int    `json:"episode_count"`
		ID           int    `json:"id"`
		PosterPath   string `json:"poster_path"`
		SeasonNumber int    `json:"season_number"`
	} `json:"seasons"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

type TVEpisode struct {
	AirDate string `json:"air_date"`
	Crew    []struct {
		ID          int    `json:"id"`
		CreditID    string `json:"credit_id"`
		Name        string `json:"name"`
		Department  string `json:"department"`
		Job         string `json:"job"`
		ProfilePath string `json:"profile_path"`
	} `json:"crew"`
	EpisodeNumber int `json:"episode_number"`
	GuestStars    []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		CreditID    string `json:"credit_id"`
		Character   string `json:"character"`
		Order       int    `json:"order"`
		ProfilePath string `json:"profile_path"`
	} `json:"guest_stars"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ID             int     `json:"id"`
	ProductionCode string  `json:"production_code"`
	SeasonNumber   int     `json:"season_number"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      int     `json:"vote_count"`
}

type TVEpisodeStills struct {
	ID     int `json:"id"`
	Stills []struct {
		AspectRatio float64     `json:"aspect_ratio"`
		FilePath    string      `json:"file_path"`
		Height      int         `json:"height"`
		Iso6391     interface{} `json:"iso_639_1"`
		VoteAverage float64     `json:"vote_average"`
		VoteCount   int         `json:"vote_count"`
		Width       int         `json:"width"`
	} `json:"stills"`
}

type ShowBackdrops struct {
	Backdrops []struct {
		AspectRatio float64     `json:"aspect_ratio"`
		FilePath    string      `json:"file_path"`
		Height      int         `json:"height"`
		Iso6391     interface{} `json:"iso_639_1"`
		VoteAverage float64     `json:"vote_average"`
		VoteCount   int         `json:"vote_count"`
		Width       int         `json:"width"`
	} `json:"backdrops"`
	ID      int `json:"id"`
	Posters []struct {
		AspectRatio float64 `json:"aspect_ratio"`
		FilePath    string  `json:"file_path"`
		Height      int     `json:"height"`
		Iso6391     string  `json:"iso_639_1"`
		VoteAverage float64 `json:"vote_average"`
		VoteCount   int     `json:"vote_count"`
		Width       int     `json:"width"`
	} `json:"posters"`
}

type cacheable struct {
	entry interface{}
}

func (c cacheable) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c.entry)
}

func (c cacheable) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &c.entry)
}

func Cacheable(m interface{}) providers.CacheableEntry {
	return &cacheable{entry: &m}
}
