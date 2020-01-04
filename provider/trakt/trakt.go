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

	provider "github.com/dpfg/kinohub-core/provider"
	"github.com/dpfg/kinohub-core/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2"
)

type Client struct {
	Config            oauth2.Config
	PreferenceStorage provider.PreferenceStorage
	Logger            *logrus.Entry
}

const (
	BaseURL = "https://api.trakt.tv"
)

func (tc *Client) AuthCodeURL() string {
	return tc.Config.AuthCodeURL("")
}

func (tc *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := tc.Config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to exchange code to token")
	}

	err = tc.PreferenceStorage.Save("trakt", token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (tc *Client) get(url string, m interface{}) error {
	t := &oauth2.Token{}
	err := tc.PreferenceStorage.Load("trakt", t)
	if err != nil {
		return errors.Wrap(err, "Unable ot load preferences")
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
		return errors.Errorf("Unexcpected response status: %s %s", resp.Status, string(body))
	}

	err = json.Unmarshal(body, m)
	if err != nil {
		return err
	}

	return nil
}

func (tc *Client) post(url string, body interface{}, response interface{}) error {
	tc.Logger.Debugf("POST to URL: %s", url)

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

	tc.Logger.Debugf("%s", bodyBytes)

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
		tc.Logger.Errorln(string(rd))
		tc.Logger.Errorln(resp.Status)
		tc.Logger.Errorln(string(respBytes))
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

func (tc *Client) TrendingShows() ([]interface{}, error) {
	m := make([]interface{}, 0)
	err := tc.get(util.JoinURL(BaseURL, "shows", "trending"), &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Settings - https://trakt.docs.apiary.io/#reference/users/settings/retrieve-settings
func (tc *Client) Settings() (interface{}, error) {
	m := make([]interface{}, 0)
	err := tc.get(util.JoinURL(BaseURL, "users", "settings", "retrieve-settings"), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (tc *Client) MyShows(from time.Time, to time.Time) ([]MyShow, error) {
	tc.Logger.Debugf("Loading My Shows: %v, %v", from, to)

	m := make([]MyShow, 0)

	fromDate := from.Format("2006-01-02")
	numDays := int(to.Sub(from).Hours() / 24)

	err := tc.get(util.JoinURL(BaseURL, "calendars", "my", "shows", fromDate, strconv.Itoa(numDays)), &m)
	if err != nil {
		tc.Logger.Error(err.Error())
		return nil, errors.WithStack(err)
	}

	return m, nil
}

// Scrobble starts scrobbling new item.
func (tc *Client) Scrobble(tmdbID int) error {
	tc.Logger.Debugf("Scrobbling %d", tmdbID)

	body := struct {
		Episode  Episode `json:"episode"`
		Progress int     `json:"progress"`
	}{
		Episode: Episode{
			Ids: EpisodeIds{
				Tmdb: tmdbID,
			},
		},
		Progress: 0,
	}

	return tc.post(util.JoinURL(BaseURL, "scrobble", "start"), body, nil)
}

// NewTraktClient creates new client
// TODO: Remove after restruct
func NewTraktClient(logger *logrus.Logger) *Client {
	return &Client{
		Config: oauth2.Config{
			ClientID:     os.Getenv("TRAKT_CLIENT_ID"),
			ClientSecret: os.Getenv("TRAKT_CLIENT_SECRET"),
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.trakt.tv/oauth/authorize",
				TokenURL: "https://api.trakt.tv/oauth/token",
			},
			RedirectURL: "http://localhost:8081/trakt/exchange",
		},
		PreferenceStorage: provider.JSONPreferenceStorage{
			Path: ".data/",
		},
		Logger: logger.WithFields(logrus.Fields{"prefix": "trakt"}),
	}
}
