package trakt

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2"
)

type TraktClient struct {
	Config            oauth2.Config
	PreferenceStorage providers.PreferenceStorage
	logger            *logrus.Logger
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

	err = json.Unmarshal(body, m)
	if err != nil {
		return err
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

func (tc *TraktClient) GetMyShows(days int64) ([]MyShow, error) {
	tc.logger.Debugln("Loading Trakt.TV My Shows")

	m := make([]MyShow, 0)

	today := "2017-08-01" //time.Now().Format("2006-01-02")
	numDays := strconv.FormatInt(days, 10)

	err := tc.get(util.JoinURL(BaseURL, "calendars", "my", "shows", today, numDays), &m)
	if err != nil {
		tc.logger.Error("Unable to load trakt.tv my shows")
		return nil, errors.WithStack(err)
	}

	return m, nil
}

func NewTraktClient() *TraktClient {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)

	return &TraktClient{
		Config: oauth2.Config{
			ClientID:     "c1bc6797965a798d9fcb83ca32c1258273c334fe543939e1378df22c1a765808",
			ClientSecret: "3cc89ccfaaf30f06c84fb87c136bb15493dc52dd1a86259ea521b31780afbb46",
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
		logger: logger,
	}
}
