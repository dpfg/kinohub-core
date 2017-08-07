package kinopub

type Item struct {
	Id      int      `json:"id,omitempty"`
	Title   string   `json:"title,omitempty"`
	Seasons []Season `json:"seasons,omitempty"`
}

type Season struct {
	Number int `json:"number,omitempty"`

	Episodes []Episode `json:"episodes,omitempty"`
}

type Episode struct {
	Title     string `json:"title,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`

	Files []File `json:"files,omitempty"`
}

type File struct {
	Quality string  `json:"quality,omitempty"`
	URL     FileURL `json:"url,omitempty"`
}

type FileURL struct {
	HTTP string `json:"http,omitempty"`
}
