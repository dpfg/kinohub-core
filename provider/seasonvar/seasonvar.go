package seasonvar

import (
	"net/url"
	"os"

	"github.com/franela/goreq"
)

// Client provides access to Seasonvar API
type Client interface {
	Search(q string) ([]interface{}, error)
}

// BaseURL points to the SeasonVar API entry point
const BaseURL = "http://api.seasonvar.ru/"

type clientImpl struct {
	apiKey string
}

func (cl clientImpl) Search(q string) ([]interface{}, error) {
	params := url.Values{}

	params.Set("command", "search")
	params.Set("key", cl.apiKey)
	params.Set("query", q)

	resp, err := goreq.Request{
		Method:      "POST",
		ContentType: "application/x-www-form-urlencoded",
		Uri:         BaseURL,
		Body:        params.Encode(),
	}.Do()

	if err != nil {
		return []interface{}{}, err
	}

	var data []interface{}
	err = resp.Body.FromJsonTo(&data)
	if err != nil {
		return nil, err
	}

	return data, nil

}

// NewClient create new instance of seasonvar client
func NewClient() Client {
	return &clientImpl{
		apiKey: os.Getenv("SV_API_KEY"),
	}
}
