package services

import (
	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/tmdb"
	"github.com/sirupsen/logrus"
)

// ContentSearch provides a way to find available media streams
type ContentBrowser interface {
	GetSeason(id, seasonNum int) (*domain.Season, error)
}

type ContentBrowserImpl struct {
	Logger  *logrus.Entry
	Kinopub kinopub.KinoPubClient
	TMDB    tmdb.Client
}

func (b ContentBrowserImpl) GetSeason(id, seasonNum int) (*domain.Season, error) {
	season, err := b.TMDB.GetTVSeason(id, seasonNum)
	if err != nil {
		return nil, err
	}

	show, err := b.TMDB.GetTVShowByID(id)
	if err != nil {
		return nil, err
	}

	ids, err := b.TMDB.GetTVShowExternalIDS(id)
	if err != nil {
		return nil, err
	}

	kpi, err := b.Kinopub.FindItemByIMDB(kinopub.StripImdbID(ids.ImdbID), show.Name)

	return &domain.Season{
		TmdbID:     season.ID,
		Name:       season.Name,
		AirDate:    season.AirDate,
		Number:     season.SeasonNumber,
		PosterPath: season.PosterPath,
		Episodes:   toDomain(season, kpi),
	}, nil
}

func toDomain(season *tmdb.TVSeason, kpi *kinopub.Item) []domain.Episode {
	r := make([]domain.Episode, 0)

	for _, episode := range season.Episodes {
		de := domain.Episode{
			Number: episode.EpisodeNumber,
			// FirstAired: episode.AirDate, // TODO:
			Overview: episode.Overview,
			Title:    episode.Name,
			TmdbID:   episode.ID,
		}

		if kpi != nil {
			if len(kpi.Seasons) >= season.SeasonNumber {
				kps := kpi.Seasons[season.SeasonNumber]

				if len(kps.Episodes) >= episode.EpisodeNumber {
					de.Files = toDomainFiles(kps.Episodes[episode.EpisodeNumber].Files)
				}
			}
		}

		r = append(r, de)
	}

	return r
}

func toDomainFiles(files []kinopub.File) []domain.File {
	r := make([]domain.File, 0)
	for _, f := range files {
		r = append(r, domain.File{
			Quality: f.Quality,
			URL:     f.URL,
		})
	}
	return r
}

func NewContentBrowser(kpc kinopub.KinoPubClient, tmdb tmdb.Client) ContentBrowser {
	return ContentBrowserImpl{
		Kinopub: kpc,
		TMDB:    tmdb,
	}
}
