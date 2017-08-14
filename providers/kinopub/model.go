package kinopub

type Item struct {
	ID       int64  `json:"id"`
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
		ID  string `json:"id"`
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
			Tracks    string `json:"tracks"`
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
			Files []struct {
				W       int    `json:"w"`
				H       int    `json:"h"`
				Quality string `json:"quality"`
				URL     struct {
					HTTP string `json:"http"`
					Hls  string `json:"hls"`
					Hls4 string `json:"hls4"`
				} `json:"url"`
			} `json:"files"`
		} `json:"episodes"`
	} `json:"seasons"`
}
