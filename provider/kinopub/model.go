package kinopub

import (
	"encoding/json"
	"strconv"

	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/util"
)

type Item struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Subtype  string `json:"subtype"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Cast     string `json:"cast"`
	Director string `json:"director"`
	Genres   []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"genres"`
	Countries []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"countries"`
	Voice    string `json:"voice"`
	Duration struct {
		Average float64 `json:"average"`
		Total   int     `json:"total"`
	} `json:"duration"`
	Langs int `json:"langs"`
	// Quality          string        `json:"quality"`
	Plot             string        `json:"plot"`
	Tracklist        []interface{} `json:"tracklist"`
	Imdb             int           `json:"imdb"`
	ImdbRating       float64       `json:"imdb_rating"`
	ImdbVotes        int           `json:"imdb_votes"`
	Kinopoisk        int           `json:"kinopoisk"`
	KinopoiskRating  float64       `json:"kinopoisk_rating"`
	KinopoiskVotes   int           `json:"kinopoisk_votes"`
	Rating           int           `json:"rating"`
	RatingVotes      string        `json:"rating_votes"`
	RatingPercentage string        `json:"rating_percentage"`
	Views            int           `json:"views"`
	Comments         int           `json:"comments"`
	Posters          struct {
		Small  string `json:"small"`
		Medium string `json:"medium"`
		Big    string `json:"big"`
	} `json:"posters"`
	Trailer struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	} `json:"trailer"`
	Finished    bool          `json:"finished"`
	Advert      bool          `json:"advert"`
	PoorQuality bool          `json:"poor_quality"`
	InWatchlist bool          `json:"in_watchlist"`
	Subscribed  bool          `json:"subscribed"`
	Subtitles   string        `json:"subtitles"`
	Bookmarks   []interface{} `json:"bookmarks"`
	Ac3         int           `json:"ac3"`
	Seasons     []struct {
		Title    string `json:"title"`
		Number   int    `json:"number"`
		Watching struct {
			Status int `json:"status"`
		} `json:"watching"`
		Episodes []struct {
			ID        int    `json:"id"`
			Title     string `json:"title"`
			Thumbnail string `json:"thumbnail"`
			Duration  int    `json:"duration"`
			Tracks    int    `json:"tracks"`
			Number    int    `json:"number"`
			Ac3       int    `json:"ac3"`
			Watched   int    `json:"watched"`
			Watching  struct {
				Status int `json:"status"`
				Time   int `json:"time"`
			} `json:"watching"`
			Subtitles []struct {
				Lang  string `json:"lang"`
				Shift int    `json:"shift"`
				Embed bool   `json:"embed"`
				URL   string `json:"url"`
			} `json:"subtitles"`
			Files []File `json:"files"`
		} `json:"episodes"`
	} `json:"seasons"`
}

func (item Item) ToDomain() *domain.Series {
	return &domain.Series{
		UID:        ToUID(item.ID),
		Overview:   item.Plot,
		PosterPath: item.Posters.Big,
		Title:      item.Title,
	}
}

type File struct {
	W       int    `json:"w"`
	H       int    `json:"h"`
	Quality string `json:"quality"`
	URL     struct {
		HTTP string `json:"http"`
		Hls  string `json:"hls"`
		Hls4 string `json:"hls4"`
	} `json:"url"`
}

func ToDomainFiles(files []File) []domain.File {
	r := make([]domain.File, 0)
	for _, f := range files {
		r = append(r, domain.File{
			Quality: f.Quality,
			URL:     f.URL,
		})
	}
	return r
}

func (item Item) MarshalBinary() (data []byte, err error) {
	return json.Marshal(item)
}

func (item *Item) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, item)
}

func (item *Item) DomainType() string {
	switch item.Type {
	case "serial":
		return domain.TypeSerial
	case "movie":
		return domain.TypeMovie
	default:
		return domain.TypeUnknown
	}
}

func (item *Item) ImdbID() string {
	sid := strconv.Itoa(item.Imdb)
	if len(sid) < 7 {
		sid = util.PadLeft(sid, "0", 7)
	}
	return "tt" + sid
}
