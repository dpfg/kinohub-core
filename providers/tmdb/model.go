package tmdb

import (
	"github.com/dpfg/kinohub-core/domain"
)

type TVShow struct {
	BackdropPath string `json:"backdrop_path"`
	CreatedBy    []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Gender      int    `json:"gender"`
		ProfilePath string `json:"profile_path"`
	} `json:"created_by"`
	EpisodeRunTime []int    `json:"episode_run_time"`
	FirstAirDate   string   `json:"first_air_date"`
	Homepage       string   `json:"homepage"`
	ID             int      `json:"id"`
	InProduction   bool     `json:"in_production"`
	Languages      []string `json:"languages"`
	LastAirDate    string   `json:"last_air_date"`
	Name           string   `json:"name"`
	Networks       []struct {
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
	Seasons     []TVSeason `json:"seasons"`
	Status      string     `json:"status"`
	Type        string     `json:"type"`
	VoteAverage float64    `json:"vote_average"`
	VoteCount   int        `json:"vote_count"`
}

func (show TVShow) ToDomain() *domain.Series {
	seasons := make([]domain.Season, 0)
	for _, season := range show.Seasons {
		seasons = append(seasons, season.ToDomain())
	}

	return &domain.Series{
		UID:        ToUID(show.ID),
		Overview:   show.Overview,
		PosterPath: ImagePath(show.PosterPath, OriginalSize),
		Title:      show.Name,
		Seasons:    seasons,
	}
}

type TVSeason struct {
	ID           int         `json:"id"`
	AirDate      string      `json:"air_date"`
	Episodes     []TVEpisode `json:"episodes"`
	Name         string      `json:"name"`
	Overview     string      `json:"overview"`
	PosterPath   string      `json:"poster_path"`
	SeasonNumber int         `json:"season_number"`
}

func (season TVSeason) ToDomain() domain.Season {
	episodes := make([]domain.Episode, 0)
	for _, episode := range season.Episodes {
		episodes = append(episodes, episode.ToDomain())
	}

	return domain.Season{
		Number:     season.SeasonNumber,
		UID:        ToUID(season.ID),
		Name:       season.Name,
		PosterPath: ImagePath(season.PosterPath, OriginalSize),
		Episodes:   episodes,
	}
}

type TVEpisode struct {
	AirDate        string  `json:"air_date"`
	EpisodeNumber  int     `json:"episode_number"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ID             int     `json:"id"`
	ProductionCode string  `json:"production_code"`
	SeasonNumber   int     `json:"season_number"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      int     `json:"vote_count"`
}

func (episode TVEpisode) ToDomain() domain.Episode {
	return domain.Episode{
		UID:       ToUID(episode.ID),
		Title:     episode.Name,
		Number:    episode.EpisodeNumber,
		Overview:  episode.Overview,
		StillPath: ImagePath(episode.StillPath, OriginalSize),
		Season:    episode.SeasonNumber,
	}
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

type Ids struct {
	ImdbID      string `json:"imdb_id"`
	FreebaseMid string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	TvdbID      int    `json:"tvdb_id"`
	TvrageID    int    `json:"tvrage_id"`
	ID          int    `json:"id"`
}

type SearchResult struct {
	TVResults []TVShow `json:"tv_results"`
}
