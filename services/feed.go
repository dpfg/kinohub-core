package services

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/tmdb"

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
	tc      *trakt.TraktClient
	kpc     kinopub.KinoPubClient
	tmdbCli tmdb.Client
	logger  *logrus.Entry
}

func (f FeedImpl) Releases(from time.Time, to time.Time) ([]FeedItem, error) {
	start := time.Now()
	m, err := f.tc.GetMyShows(from, to)
	if err != nil {
		return nil, err
	}
	f.logger.Debugf("Loaded recent shows in %s", time.Now().Sub(start).String())

	r := make([]FeedItem, 0)
	for _, item := range m {
		imdbID, _ := strconv.Atoi(strings.TrimLeft(item.Show.Ids.Imdb, "tt"))
		ep, err := f.kpc.GetEpisode(imdbID, item.Show.Title, item.Episode.Season, item.Episode.Number)
		if err != nil {
			f.logger.Errorln(errors.WithMessage(err, "Cannot load KinHub episode").Error())
			continue
		}

		r = append(r, FeedItem{
			ContentAvailable: ep != nil,
			Show: domain.Show{
				Title: item.Show.Title,
				UID:   tmdb.ToUID(item.Show.Ids.Tmdb),
			},
			Episode: domain.Episode{
				UID:        tmdb.ToUID(item.Episode.Ids.Tmdb),
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
		logger: logger.WithFields(logrus.Fields{"prefix": "feed"}),
	}
}
