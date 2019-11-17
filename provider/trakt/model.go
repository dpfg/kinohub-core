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

// Settings - https://trakt.docs.apiary.io/#reference/users/settings/retrieve-settings
type Settings struct {
	User struct {
		Username string `json:"username"`
		Private  bool   `json:"private"`
		Name     string `json:"name"`
		Vip      bool   `json:"vip"`
		VipEp    bool   `json:"vip_ep"`
		Ids      struct {
			Slug string `json:"slug"`
		} `json:"ids"`
		JoinedAt time.Time `json:"joined_at"`
		Location string    `json:"location"`
		About    string    `json:"about"`
		Gender   string    `json:"gender"`
		Age      int       `json:"age"`
		Images   struct {
			Avatar struct {
				Full string `json:"full"`
			} `json:"avatar"`
		} `json:"images"`
		VipOg    bool `json:"vip_og"`
		VipYears int  `json:"vip_years"`
	} `json:"user"`
	Account struct {
		Timezone   string `json:"timezone"`
		DateFormat string `json:"date_format"`
		Time24Hr   bool   `json:"time_24hr"`
		CoverImage string `json:"cover_image"`
	} `json:"account"`
	Connections struct {
		Facebook bool `json:"facebook"`
		Twitter  bool `json:"twitter"`
		Google   bool `json:"google"`
		Tumblr   bool `json:"tumblr"`
		Medium   bool `json:"medium"`
		Slack    bool `json:"slack"`
	} `json:"connections"`
	SharingText struct {
		Watching string `json:"watching"`
		Watched  string `json:"watched"`
		Rated    string `json:"rated"`
	} `json:"sharing_text"`
}
