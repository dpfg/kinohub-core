package trakt

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2"
)

type TraktClient struct {
	Config            oauth2.Config
	PreferenceStorage providers.PreferenceStorage
	logger            *logrus.Entry
}

const (
	BaseURL = "https://api.trakt.tv"
)

func (tc *TraktClient) GetAuthCodeURL() string {
	return tc.Config.AuthCodeURL("", oauth2.AccessTypeOffline)
}

func (tc *TraktClient) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return tc.Config.Exchange(ctx, code)
}

func (tc *TraktClient) get(url string, m interface{}) error {
	t := &oauth2.Token{}
	err := tc.PreferenceStorage.Load("trakt", t)
	if err != nil {
		return err
	}

	cl := tc.Config.Client(context.TODO(), t)
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("trakt-api-key", tc.Config.ClientID)

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		rd, _ := httputil.DumpRequest(req, false)
		tc.logger.Errorln(string(rd))
		tc.logger.Errorln(resp.Status)
		tc.logger.Error(string(body))
		return nil
	}

	err = json.Unmarshal(body, m)
	if err != nil {
		return err
	}

	return nil
}

func (tc *TraktClient) post(url string, body interface{}, response interface{}) error {
	tc.logger.Debugf("POST to URL: %s", url)

	t := &oauth2.Token{}
	err := tc.PreferenceStorage.Load("trakt", t)
	if err != nil {
		return err
	}

	cl := tc.Config.Client(context.TODO(), t)

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	tc.logger.Debugf("%s", bodyBytes)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))

	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("trakt-api-key", tc.Config.ClientID)

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		rd, _ := httputil.DumpRequest(req, false)
		tc.logger.Errorln(string(rd))
		tc.logger.Errorln(resp.Status)
		tc.logger.Errorln(string(respBytes))
		return nil
	}

	if response != nil {
		err = json.Unmarshal(respBytes, response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *TraktClient) GetTrendingShows() ([]interface{}, error) {
	m := make([]interface{}, 0)
	err := tc.get(util.JoinURL(BaseURL, "shows", "trending"), &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (tc *TraktClient) GetMyShows(from time.Time, to time.Time) ([]MyShow, error) {
	tc.logger.Debugf("Loading My Shows: %v, %v", from, to)

	m := make([]MyShow, 0)

	fromDate := from.Format("2006-01-02")
	numDays := int(to.Sub(from).Hours() / 24)

	err := tc.get(util.JoinURL(BaseURL, "calendars", "my", "shows", fromDate, strconv.Itoa(numDays)), &m)
	if err != nil {
		tc.logger.Error(err.Error())
		return nil, errors.WithStack(err)
	}

	return m, nil
}

// Scrobble starts scrobbling new item.
func (tc *TraktClient) Scrobble(imdbId string) error {
	tc.logger.Debugf("Scrobbling %d", imdbId)

	body := struct {
		Episode  Episode `json:"episode"`
		Progress int     `json:"progress"`
	}{
		Episode: Episode{
			Ids: EpisodeIds{
				Imdb: imdbId,
			},
		},
		Progress: 0,
	}

	return tc.post(util.JoinURL(BaseURL, "scrobble", "start"), body, nil)
}

func NewTraktClient(logger *logrus.Logger) *TraktClient {
	return &TraktClient{
		Config: oauth2.Config{
			ClientID:     os.Getenv("TRAKT_CLIENT_ID"),
			ClientSecret: os.Getenv("TRAKT_CLIENT_SECRET"),
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.trakt.tv/oauth/authorize",
				TokenURL: "https://api.trakt.tv/oauth/token",
			},
			RedirectURL: "http://localhost:1323/callback/trakt",
		},
		PreferenceStorage: providers.JSONPreferenceStorage{
			Path: ".data/",
		},
		logger: logger.WithFields(logrus.Fields{"prefix": "trakt"}),
	}
}
