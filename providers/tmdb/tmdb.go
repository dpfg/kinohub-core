package tmdb

import (
	"net/http"
	"os"
	"strconv"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/util"
	"github.com/franela/goreq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client is an interface that describes capabilities of TMDB API.
type Client interface {
	// Get the primary TV show details by id.
	GetTVShowByID(id int) (*TVShow, error)
	// Get the external ids for a TV show
	GetTVShowExternalIDS(id int)
	// Get the images that belong to a TV show.
	GetTVShowImages(id int) (*ShowBackdrops, error)

	// Get the TV episode details by id.
	GetTVEpisode(tvID int, seasonNum int, episodeNum int) (*TVEpisode, error)
	// Get the images that belong to a TV episode.
	GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int) (*TVEpisodeStills, error)
}

const (
	// BaseURL is the TMDB API base url
	BaseURL = "https://api.themoviedb.org/3"
)

// ClientImpl is a default implementation of TMDB API consumer
type ClientImpl struct {
	apiKey            string
	logger            *logrus.Entry
	cache             providers.CacheFactory
	preferenceStorage providers.PreferenceStorage
}

func (cl ClientImpl) doGet(url string, body interface{}) error {
	// cache := cl.cache.Get("TMDB_ENTITIES", time.Hour*24)

	// cacheEntry := &struct {
	// 	ID   int
	// 	body interface{}
	// }{
	// 	ID:   -1,
	// 	body: body,
	// }

	// cacheKey := url

	// err := cache.Load(cacheKey, cacheEntry)
	// if err != nil {
	// 	return err
	// }

	// cl.logger.Debugln(cacheEntry.body)
	// if cacheEntry.ID != -1 {
	// 	return nil
	// }

	resp, err := goreq.Request{
		Method: "GET",
		Uri:    url,
		QueryString: struct {
			APIKey string `url:"api_key,omitempty"`
		}{
			APIKey: cl.apiKey,
		},
	}.Do()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Network error - %s", resp.Status)
	}

	err = resp.Body.FromJsonTo(body)
	if err != nil {
		return err
	}

	// cache.Save(cacheKey, &struct {
	// 	ID   int
	// 	body interface{}
	// }{
	// 	ID:   1,
	// 	body: body,
	// })

	return nil
}

// GetTVShowByID returns the primary TV show details by id.
func (cl ClientImpl) GetTVShowByID(id int) (*TVShow, error) {
	cl.logger.Debugf("Getting TMDB show by ID=[%d]", id)

	show := &TVShow{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", strconv.Itoa(id)), show)
	if err != nil {
		return nil, err
	}

	return show, nil
}

// GetTVShowExternalIDS returns the external ids for a TV show
func (cl ClientImpl) GetTVShowExternalIDS(id int) {
	panic("not implemented")
}

// GetTVShowImages returns the images that belong to a TV show.
func (cl ClientImpl) GetTVShowImages(id int) (*ShowBackdrops, error) {
	backdrops := &ShowBackdrops{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", id, "images"), backdrops)
	if err != nil {
		return nil, err
	}

	return backdrops, nil
}

// GetTVEpisode returns TV episode details by id.
func (cl ClientImpl) GetTVEpisode(tvID int, seasonNum int, episodeNum int) (*TVEpisode, error) {
	url := util.JoinURL(BaseURL, "tv", tvID, "season", seasonNum, "episode", episodeNum)

	episode := &TVEpisode{}
	err := cl.doGet(url, episode)

	if err != nil {
		return nil, err
	}

	return episode, nil
}

// GetTVEpisodeImages returnes the images that belong to a TV episode.
func (cl ClientImpl) GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int) (*TVEpisodeStills, error) {
	url := util.JoinURL(BaseURL, "tv", tvID, "season", seasonNum, "episode", episodeNum, "images")

	stills := &TVEpisodeStills{}
	err := cl.doGet(url, stills)

	if err != nil {
		return nil, err
	}

	return stills, nil
}

// New returns new TMDB API client
func New(logger *logrus.Logger, cf providers.CacheFactory, ps providers.PreferenceStorage) Client {
	return ClientImpl{
		apiKey:            os.Getenv("TMDB_API_KEY"),
		preferenceStorage: ps,
		cache:             cf,
		logger:            logger.WithField("prefix", "tmdb"),
	}
}
