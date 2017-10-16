package services

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dpfg/kinohub-core/providers/kinopub"

	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/providers/trakt"
)

type Feed interface {
	Releases(from time.Time, to time.Time) ([]FeedItem, error)
}

type FeedItem struct {
	Show             domain.Show    `json:"show,omitempty"`
	Episode          domain.Episode `json:"episode,omitempty"`
	ContentAvailable bool           `json:"content_available,omitempty"`
}

type FeedImpl struct {
	tc     *trakt.TraktClient
	kpc    kinopub.KinoPubClient
	logger *logrus.Entry
}

func (f FeedImpl) Releases(from time.Time, to time.Time) ([]FeedItem, error) {
	m, err := f.tc.GetMyShows(from, to)
	if err != nil {
		return nil, err
	}

	r := make([]FeedItem, 0)
	for _, item := range m {
		id, _ := strconv.Atoi(strings.TrimLeft(item.Show.Ids.Imdb, "tt"))
		ep, err := f.kpc.GetEpisode(id, item.Show.Title, item.Episode.Season, item.Episode.Number)
		if err != nil {
			return nil, err
		}

		r = append(r, FeedItem{
			ContentAvailable: ep != nil,
			Show: domain.Show{
				Title:  item.Show.Title,
				ImdbID: item.Show.Ids.Imdb,
			},
			Episode: domain.Episode{
				ImdbID:     item.Episode.Ids.Imdb,
				Title:      item.Episode.Title,
				Number:     item.Episode.Number,
				Season:     item.Episode.Season,
				FirstAired: item.FirstAired,
			},
		})
	}
	return r, nil
}

func NewFeed(tc *trakt.TraktClient, kpc kinopub.KinoPubClient, logger *logrus.Logger) Feed {
	return FeedImpl{
		tc:     tc,
		kpc:    kpc,
		logger: logger.WithFields(logrus.Fields{"prefix": "kinpub"}),
	}
}
