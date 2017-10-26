package tmdb

import (
	"net/http"
	"os"
	"strconv"
	"time"

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

	//
	GetTVSeason(id, seasonNum int) (*TVSeason, error)

	// Get the TV episode details by id.
	GetTVEpisode(tvID int, seasonNum int, episodeNum int) (*TVEpisode, error)
	// Get the images that belong to a TV episode.
	GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int) (*TVEpisodeStills, error)
}

const (
	// BaseURL is the TMDB API base url
	BaseURL    = "https://api.themoviedb.org/3"
	ImgBaseURL = "https://image.tmdb.org/t/p/"
)

// ClientImpl is a default implementation of TMDB API consumer
type ClientImpl struct {
	apiKey            string
	logger            *logrus.Entry
	cache             providers.CacheFactory
	preferenceStorage providers.PreferenceStorage
}

func (cl ClientImpl) doGet(url string, body providers.CacheableEntry) error {
	cache := cl.cache.Get("TMDB_ENTITIES", time.Hour*24)

	if cache.Load(url, body) {
		return nil
	}

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

	rb, err := resp.Body.ToString()
	if err != nil {
		return errors.WithMessage(err, "cannot read response body")
	}

	err = body.UnmarshalBinary([]byte(rb))
	if err != nil {
		return errors.WithMessage(err, "cannot unmarshal response")
	}

	cache.Save(url, body)

	return nil
}

// GetTVShowByID returns the primary TV show details by id.
func (cl ClientImpl) GetTVShowByID(id int) (*TVShow, error) {
	cl.logger.Debugf("Getting TMDB show by ID=[%d]", id)

	show := &TVShow{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", strconv.Itoa(id)), Cacheable(show))
	if err != nil {
		return nil, err
	}

	cl.logger.Debugf("TMDB show ID=[%d] has been loaded", id)

	return show, nil
}

// GetTVShowExternalIDS returns the external ids for a TV show
func (cl ClientImpl) GetTVShowExternalIDS(id int) {
	panic("not implemented")
}

// GetTVShowImages returns the images that belong to a TV show.
func (cl ClientImpl) GetTVShowImages(id int) (*ShowBackdrops, error) {
	backdrops := &ShowBackdrops{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", id, "images"), Cacheable(backdrops))
	if err != nil {
		return nil, err
	}

	return backdrops, nil
}

// GetTVSeason return the detailed information about the season
func (cl ClientImpl) GetTVSeason(id, seasonNum int) (*TVSeason, error) {
	season := &TVSeason{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", id, "season", seasonNum), Cacheable(season))
	if err != nil {
		return nil, err
	}

	season.PosterPath = FullImagePath(-1, season.PosterPath)

	return season, nil
}

// GetTVEpisode returns TV episode details by id.
func (cl ClientImpl) GetTVEpisode(tvID int, seasonNum int, episodeNum int) (*TVEpisode, error) {
	url := util.JoinURL(BaseURL, "tv", tvID, "season", seasonNum, "episode", episodeNum)

	episode := &TVEpisode{}
	err := cl.doGet(url, Cacheable(episode))

	if err != nil {
		return nil, err
	}

	return episode, nil
}

// GetTVEpisodeImages returnes the images that belong to a TV episode.
func (cl ClientImpl) GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int) (*TVEpisodeStills, error) {
	url := util.JoinURL(BaseURL, "tv", tvID, "season", seasonNum, "episode", episodeNum, "images")

	stills := &TVEpisodeStills{}
	err := cl.doGet(url, Cacheable(stills))

	if err != nil {
		return nil, err
	}

	return stills, nil
}

func FullImagePath(w int, path string) string {
	if w > 600 {
		panic("unsupported image size")
	}

	size := strconv.Itoa(w)
	if w < 0 {
		size = "original"
	}

	return ImgBaseURL + "/" + size + path
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
