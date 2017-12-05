package services

import (
	"sort"
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
	Show             domain.Series  `json:"show,omitempty"`
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

	m, err := f.tc.GetMyShows(from, to)
	if err != nil {
		return nil, err
	}

	sort.Slice(m, func(i, j int) bool { return m[i].FirstAired.After(m[j].FirstAired) })

	r := make([]FeedItem, 0)
	for _, item := range m {

		imdbID, _ := strconv.Atoi(strings.TrimLeft(item.Show.Ids.Imdb, "tt"))
		ep, err := f.kpc.GetEpisode(imdbID, item.Show.Title, item.Episode.Season, item.Episode.Number)
		if err != nil {
			f.logger.Errorln(errors.WithMessage(err, "Cannot load KinHub episode").Error())
			continue
		}

		images, _ := f.tmdbCli.GetTVEpisodeImages(item.Show.Ids.Tmdb, item.Episode.Season, item.Episode.Number)
		episodeStill := ""
		if len(images.Stills) > 0 {
			episodeStill = tmdb.ImagePath(images.Stills[0].FilePath, 320)
		}

		r = append(r, FeedItem{
			ContentAvailable: ep != nil,
			Show: domain.Series{
				Title: item.Show.Title,
				UID:   tmdb.ToUID(item.Show.Ids.Tmdb),
			},
			Episode: domain.Episode{
				UID:        tmdb.ToUID(item.Episode.Ids.Tmdb),
				Title:      item.Episode.Title,
				Number:     item.Episode.Number,
				Season:     item.Episode.Season,
				FirstAired: item.FirstAired,
				StillPath:  episodeStill,
			},
		})
	}

	return r, nil
}

func NewFeed(tc *trakt.TraktClient, kpc kinopub.KinoPubClient, tmdb tmdb.Client, logger *logrus.Logger) Feed {
	return FeedImpl{
		tc:      tc,
		kpc:     kpc,
		tmdbCli: tmdb,
		logger:  logger.WithFields(logrus.Fields{"prefix": "feed"}),
	}
}
