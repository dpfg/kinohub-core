package kinopub

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/franela/goreq"
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
}

const (
	BaseURL  = "https://api.service-kp.com/v1/"
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
	log.Println("refreshing token....")
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
		return nil, err
	}

	resp, err := goreq.Request{
		Method: "GET",
		Uri:    BaseURL + "items",
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
		return nil, err
	}

	return m.Items, nil
}

func (cl KinoPubClientImpl) GetItemById(id int) (*Item, error) {
	t, _ := cl.getToken()

	resp, err := goreq.Request{
		Method: "GET",
		Uri:    path.Join(BaseURL, "item", string(id)),
		QueryString: authQuery{
			AccessToken: t.AccessToken,
		},
	}.Do()

	if err != nil {
		return nil, err
	}

	item := &Item{}
	err = resp.Body.FromJsonTo(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}
