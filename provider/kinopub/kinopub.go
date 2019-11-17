package kinopub

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"strings"

	provider "github.com/dpfg/kinohub-core/provider"
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

	GetEpisode(imdbID int, title string, seasonNum int, episodeNum int) (interface{}, error)

	FindItemByIMDB(imdbID int, title string) (*Item, error)
}

type ItemsFilter struct {
	Title string `url:"title,omitempty"`
}

type KinoPubClientImpl struct {
	ClientID          string
	ClientSecret      string
	PreferenceStorage provider.PreferenceStorage
	CacheFactory      provider.CacheFactory
	Logger            *logrus.Entry
	fixer             *fixer
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

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Cannot refresh kinopub token: service response - %s", resp.Status)
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

	if len(nt.RefreshToken) == 0 || len(nt.AccessToken) == 0 {
		return errors.Errorf("Cannot refresh kinohub token: empty refresh token.")
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

func (cl KinoPubClientImpl) GetItemById(id int) (*Item, error) {
	cl.Logger.Debugf("Loading kinpub item by ID=%d", id)

	cache := cl.CacheFactory.Get("KP_GetItemById", time.Hour)
	cacheKey := fmt.Sprint(id)

	item := &Item{}

	if cache.Load(cacheKey, item) {
		// cl.fixer.fixID(item)
		return item, nil
	}

	t, err := cl.getToken()
	if err != nil {
		return nil, errors.Wrap(err, "No auth")
	}

	cl.Logger.Debugln("Fetching kinpub item from the remote service")
	resp, err := goreq.Request{
		Method: "GET",
		Uri:    util.JoinURL(BaseURL, "items", strconv.FormatInt(int64(id), 10)),
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

	// cl.fixer.fixID(&m.Item)

	cache.Save(cacheKey, &m.Item)

	return &m.Item, nil
}

// FindItemByIMDB search item by IMDB id. As there is no filter data by id, getch by title and then filter manually.
func (cl KinoPubClientImpl) FindItemByIMDB(imdbID int, title string) (*Item, error) {
	cache := cl.CacheFactory.Get("KP_FindItemByIMDB", time.Hour*24*7)
	cacheKey := strconv.Itoa(imdbID)

	item := &Item{}

	if cache.Load(cacheKey, item) {
		return item, nil
	}

	title = truncateProblematicTitle(title)
	cl.Logger.Debugf("Searching item by IMDB Id [%d, %s] on remote host.", imdbID, title)
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

	it, err := cl.GetItemById(int(item.ID))
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

// NewKinoPubClient returns new kinopub client
func NewKinoPubClient(logger *logrus.Logger, cf provider.CacheFactory) KinoPubClient {
	return KinoPubClientImpl{
		ClientID:     os.Getenv("KINOPUB_CLIENT_ID"),
		ClientSecret: os.Getenv("KINOPUB_CLIENT_SECRET"),
		PreferenceStorage: provider.JSONPreferenceStorage{
			Path: ".data/",
		},
		CacheFactory: cf,
		Logger:       logger.WithFields(logrus.Fields{"prefix": "kinpub"}),
		fixer:        &fixer{logger: logger.WithFields(logrus.Fields{"prefix": "id-fixer"})},
	}
}

func StripImdbID(id string) int {
	i, _ := strconv.Atoi(strings.Replace(id, "t", "", -1))
	return i
}

func ToImdbID(id int) string {
	return fmt.Sprintf("tt%d", id)
}

func ToUID(id int) string {
	return fmt.Sprintf("%s%d", provider.IDTypeKinoHub, id)
}

func ParseUID(uid string) (int, error) {
	if !strings.HasPrefix(uid, provider.IDTypeKinoHub) {
		return -1, errors.New("Invalid UID type")
	}

	return strconv.Atoi(strings.TrimLeft(uid, provider.IDTypeKinoHub))
}

func truncateProblematicTitle(title string) string {
	if strings.HasPrefix(title, "Marvel's") {
		return strings.TrimPrefix(title, "Marvel's")
	}
	return title
}
