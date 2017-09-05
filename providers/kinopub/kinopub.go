package kinopub

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/util"
	"github.com/franela/goreq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Token struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

func (t *Token) IsValid() bool {
	if t.AccessToken == "" || t.RefreshToken == "" || t.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

type KinoPubClient interface {
	SearchItemBy(q ItemsFilter) ([]Item, error)

	GetItemById(id int) (*Item, error)
}

type ItemsFilter struct {
	Title string `url:"title,omitempty"`
}

type KinoPubClientImpl struct {
	ClientID          string
	ClientSecret      string
	PreferenceStorage providers.PreferenceStorage
	CacheFactory      providers.CacheFactory
	Logger            *logrus.Logger
}

const (
	BaseURL  = "https://api.service-kp.com/v1"
	TokenURL = "https://api.service-kp.com/oauth2/token"

	KinoPubPrefKey = "kinopub"
)

type authQuery struct {
	AccessToken string `url:"access_token"`
}

func (cl KinoPubClientImpl) getToken() (*Token, error) {
	t := &Token{}
	err := cl.PreferenceStorage.Load(KinoPubPrefKey, t)
	if err != nil {
		return nil, err
	}

	if !t.IsValid() {
		if err = cl.refreshToken(t); err != nil {
			return nil, err
		}

		if err = cl.PreferenceStorage.Load(KinoPubPrefKey, t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (cl KinoPubClientImpl) refreshToken(t *Token) error {
	cl.Logger.Debugln("refreshing token....")
	resp, err := goreq.Request{
		Method: "POST",
		Uri:    TokenURL,
		QueryString: struct {
			GrantType    string `url:"grant_type,omitempty"`
			ClientID     string `url:"client_id,omitempty"`
			ClientSecret string `url:"client_secret,omitempty"`
			RefreshToken string `url:"refresh_token,omitempty"`
		}{
			GrantType:    "refresh_token",
			ClientID:     cl.ClientID,
			ClientSecret: cl.ClientSecret,
			RefreshToken: t.RefreshToken,
		},
	}.Do()

	if err != nil {
		return err
	}

	nt := &struct {
		AccessToken  string `json:"access_token,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresIn    int64  `json:"expires_in,omitempty"`
	}{}
	err = resp.Body.FromJsonTo(nt)
	if err != nil {
		return err
	}

	cl.PreferenceStorage.Save(KinoPubPrefKey, &Token{
		AccessToken:  nt.AccessToken,
		RefreshToken: nt.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(nt.ExpiresIn) * time.Second),
	})
	return nil
}

func (cl KinoPubClientImpl) SearchItemBy(q ItemsFilter) ([]Item, error) {
	t, err := cl.getToken()
	if err != nil {
		return nil, errors.Wrap(err, "No auth")
	}

	resp, err := goreq.Request{
		Method: "GET",
		Uri:    util.JoinURL(BaseURL, "items"),
		QueryString: struct {
			Title       string `url:"title,omitempty"`
			AccessToken string `url:"access_token,omitempty"`
		}{
			Title:       q.Title,
			AccessToken: t.AccessToken,
		},
	}.Do()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %s", resp.Status)
	}

	m := &struct {
		Items []Item `json:"items,omitempty"`
	}{}

	err = resp.Body.FromJsonTo(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return m.Items, nil
}

func (cl KinoPubClientImpl) GetItemById(id int64) (*Item, error) {
	cl.Logger.Debugf("Loading kinpub item by ID=%d", id)

	cache := cl.CacheFactory.Get("KP_GetItemById", time.Hour)
	cacheKey := fmt.Sprint(id)
	item := &Item{ID: -1}
	err := cache.Load(cacheKey, item)

	if err != nil {
		return nil, errors.WithMessage(err, "Failed to get item from the cache")
	}

	if item.ID != -1 {
		cl.Logger.Debugln("Kinpub item has been loaded using cache")
		return item, nil
	}

	t, err := cl.getToken()
	if err != nil {
		return nil, errors.Wrap(err, "No auth")
	}

	cl.Logger.Debugln("Fetching kinpub item from the remote service")
	resp, err := goreq.Request{
		Method: "GET",
		Uri:    util.JoinURL(BaseURL, "items", strconv.FormatInt(id, 10)),
		QueryString: struct {
			AccessToken string `url:"access_token,omitempty"`
		}{
			AccessToken: t.AccessToken,
		},
	}.Do()

	if err != nil {
		return nil, errors.WithMessage(err, "Can't fetch item")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %s", resp.Status)
	}

	m := &struct {
		Item Item `json:"item,omitempty"`
	}{}

	err = resp.Body.FromJsonTo(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = cache.Save(cacheKey, m.Item)
	if err != nil {
		return nil, errors.WithMessage(err, "Can't save item to the cache")
	}

	return &m.Item, nil
}

// FindItemByIMDB search item by IMDB id. As there is no filter data by id, getch by title and then filter manually.
func (cl KinoPubClientImpl) FindItemByIMDB(imdbID int, title string) (*Item, error) {
	cache := cl.CacheFactory.Get("KP_FindItemByIMDB", time.Hour*24*7)
	cacheKey := strconv.Itoa(imdbID)
	item := &Item{ID: -1}
	err := cache.Load(cacheKey, item)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if item.ID != -1 {
		cl.Logger.Debugln("Found item by IMDB in cache")
		return item, nil
	}

	cl.Logger.Debugln("Searching item by IMDB Id on remote host.")
	items, err := cl.SearchItemBy(ItemsFilter{
		Title: title,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "Can't find item")
	}

	for _, item := range items {
		if item.Imdb == imdbID {
			cache.Save(cacheKey, &item)
			return &item, nil
		}
	}

	return nil, nil
}

// GetEpisode returns kinopub episode structure by season number (1-based) and episode number (1-based)
func (cl KinoPubClientImpl) GetEpisode(imdbID int, title string, seasonNum int, episodeNum int) (interface{}, error) {
	item, err := cl.FindItemByIMDB(imdbID, title)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, nil
	}

	cl.Logger.Debugf("Found %s. Query: %s", title, item.Title)

	it, err := cl.GetItemById(item.ID)
	if err != nil {
		return nil, errors.WithMessage(err, "Can't load kinopub item by id")
	}

	cl.Logger.Debugf("Kinpub Item %d has been loaded", item.ID)
	for _, season := range it.Seasons {
		if season.Number != seasonNum {
			continue
		}

		for _, episode := range season.Episodes {
			if episode.Number == episodeNum {
				return episode, nil
			}
		}
	}
	return nil, nil
}
