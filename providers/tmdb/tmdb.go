package tmdb

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	GetTVShowExternalIDS(id int) (*Ids, error)
	// Get the images that belong to a TV show.
	GetTVShowImages(id int) (*ShowBackdrops, error)
	//
	GetTVSeason(id, seasonNum int) (*TVSeason, error)
	// Get the TV episode details by id.
	GetTVEpisode(tvID int, seasonNum int, episodeNum int) (*TVEpisode, error)
	// Get the images that belong to a TV episode.
	GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int) (TVEpisodeStills, error)

	FindByExternalID(id string) (*SearchResult, error)

	FindTVShowByExternalID(id string) (*TVShow, error)
}

const (
	// BaseURL is the TMDB API base url
	BaseURL = "https://api.themoviedb.org/3"

	// ImgBaseURL is the base path to images
	ImgBaseURL = "https://image.tmdb.org/t/p/"
)

// ClientImpl is a default implementation of TMDB API consumer
type ClientImpl struct {
	apiKey            string
	logger            *logrus.Entry
	cache             providers.CacheFactory
	preferenceStorage providers.PreferenceStorage
}

func (cl ClientImpl) doGet(uri string, qp url.Values, body providers.CacheEntry) error {
	cache := cl.cache.Get("TMDB_ENTITIES", 24*time.Hour)

	if cache.Load(uri, body) {
		return nil
	}

	if qp == nil {
		qp = url.Values{}
	}

	qp.Add("api_key", cl.apiKey)

	resp, err := goreq.Request{
		Method:      "GET",
		Uri:         uri,
		QueryString: qp,
	}.Do()

	if resp == nil {
		return errors.New("Network error: response is null")
	}

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

	cache.Save(uri, body)

	return nil
}

// GetTVShowByID returns the primary TV show details by id.
func (cl ClientImpl) GetTVShowByID(id int) (*TVShow, error) {
	cl.logger.Debugf("Getting TMDB show by ID=[%d]", id)

	show := &TVShow{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", strconv.Itoa(id)), nil, providers.Cacheable(show))
	if err != nil {
		return nil, err
	}

	cl.logger.Debugf("TMDB show ID=[%d] has been loaded", id)

	return show, nil
}

// GetTVShowExternalIDS returns the external ids for a TV show
func (cl ClientImpl) GetTVShowExternalIDS(id int) (*Ids, error) {
	ids := &Ids{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", id, "external_ids"), nil, providers.Cacheable(ids))
	if err != nil {
		return nil, err
	}

	return ids, err
}

// GetTVShowImages returns the images that belong to a TV show.
func (cl ClientImpl) GetTVShowImages(id int) (*ShowBackdrops, error) {
	backdrops := &ShowBackdrops{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", id, "images"), nil, providers.Cacheable(backdrops))
	if err != nil {
		return nil, err
	}

	return backdrops, nil
}

// GetTVSeason return the detailed information about the season
func (cl ClientImpl) GetTVSeason(seriesID, seasonNum int) (*TVSeason, error) {
	season := &TVSeason{}
	err := cl.doGet(util.JoinURL(BaseURL, "tv", seriesID, "season", seasonNum), nil, providers.Cacheable(season))
	if err != nil {
		return nil, err
	}

	season.PosterPath = ImagePath(season.PosterPath, OriginalSize)

	return season, nil
}

// GetTVEpisode returns TV episode details by id.
func (cl ClientImpl) GetTVEpisode(tvID int, seasonNum int, episodeNum int) (*TVEpisode, error) {
	url := util.JoinURL(BaseURL, "tv", tvID, "season", seasonNum, "episode", episodeNum)

	episode := &TVEpisode{}
	err := cl.doGet(url, nil, providers.Cacheable(episode))

	if err != nil {
		return nil, err
	}

	return episode, nil
}

// GetTVEpisodeImages returnes the images that belong to a TV episode.
func (cl ClientImpl) GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int) (TVEpisodeStills, error) {
	url := util.JoinURL(BaseURL, "tv", tvID, "season", seasonNum, "episode", episodeNum, "images")

	stills := TVEpisodeStills{}
	err := cl.doGet(url, nil, providers.Cacheable(&stills))

	if err != nil {
		return stills, err
	}

	return stills, nil
}

// FindByExternalID search TMDB entry by IMDB id
func (cl ClientImpl) FindByExternalID(id string) (*SearchResult, error) {
	uri := util.JoinURL(BaseURL, "find", id)
	result := &SearchResult{}

	err := cl.doGet(
		uri,
		map[string][]string{"external_source": []string{"imdb_id"}},
		providers.Cacheable(result),
	)
	if err != nil {
		return nil, err
	}
	cl.logger.Debugf("Search by external id: tv=%d", len(result.TVResults))

	return result, nil
}

func (cl ClientImpl) FindTVShowByExternalID(id string) (*TVShow, error) {
	result, err := cl.FindByExternalID(id)
	if err != nil {
		return nil, err
	}

	if len(result.TVResults) > 0 {
		return &result.TVResults[0], nil
	}

	return nil, nil
}

// OriginalSize is a parameter to ImagePath to get url to image in original size
const OriginalSize = -1

// ImagePath returns absolute URL to the image with specified width
func ImagePath(tmdbPath string, w int) string {
	if len(tmdbPath) == 0 {
		return ""
	}

	if w > 600 {
		panic("unsupported image size")
	}

	size := "w" + strconv.Itoa(w)
	if w == OriginalSize {
		size = "original"
	}

	return ImgBaseURL + size + tmdbPath
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

func ToUID(id int) string {
	return fmt.Sprintf("%s%d", providers.ID_TYPE_TMDB, id)
}

func ParseUID(uid string) (int, error) {
	if !strings.HasPrefix(uid, providers.ID_TYPE_TMDB) {
		return -1, errors.New("Invalid UID type")
	}

	return strconv.Atoi(strings.TrimLeft(uid, providers.ID_TYPE_TMDB))
}
