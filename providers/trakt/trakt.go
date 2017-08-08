package trakt

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type TraktClient struct {
	config oauth2.Config
}

func (tc *TraktClient) GetAuthCodeURL() string {
	return tc.config.AuthCodeURL("", oauth2.AccessTypeOffline)
}

func (tc *TraktClient) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return tc.config.Exchange(ctx, code)
}

func NewTraktClient() *TraktClient {
	return &TraktClient{
		config: oauth2.Config{
			ClientID:     "c1bc6797965a798d9fcb83ca32c1258273c334fe543939e1378df22c1a765808",
			ClientSecret: "3cc89ccfaaf30f06c84fb87c136bb15493dc52dd1a86259ea521b31780afbb46",
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.trakt.tv/oauth/authorize",
				TokenURL: "https://api.trakt.tv/oauth/token",
			},
			RedirectURL: "http://localhost:1323/callback/trakt",
		},
	}
}

// traktTest uses OAuth2 to test connection
func traktTest() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "c1bc6797965a798d9fcb83ca32c1258273c334fe543939e1378df22c1a765808",
		ClientSecret: "3cc89ccfaaf30f06c84fb87c136bb15493dc52dd1a86259ea521b31780afbb46",
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.trakt.tv/oauth/authorize",
			TokenURL: "https://api.trakt.tv/oauth/token",
		},
		RedirectURL: "http://mbp-ac.local:8080/trakt",
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	givenURL := conf.AuthCodeURL("", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", givenURL)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	// saveToken(tok)

	fmt.Println(tok.Expiry.Format(time.RFC3339))
	client := conf.Client(ctx, tok)
	req, _ := http.NewRequest("GET", "https://api.trakt.tv/shows/trending", nil)

	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("trakt-api-key", conf.ClientID)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Something went wrong :(")
		fmt.Println(err.Error())
		return
	}

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	for k, v := range resp.Header {
		fmt.Printf("%v - %v\n", k, v)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(body))
}
