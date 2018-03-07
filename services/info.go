package services

import (
	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/tmdb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContentSearch provides a way to find available media streams
type ContentBrowser interface {
	GetShow(uid string) (*domain.Series, error)
	GetSeason(uid string, seasonNum int) (*domain.Season, error)
}

type ContentBrowserImpl struct {
	Logger  *logrus.Entry
	Kinopub kinopub.KinoPubClient
	TMDB    tmdb.Client
}

func (b ContentBrowserImpl) GetSeason(uid string, seasonNum int) (*domain.Season, error) {
	if !providers.MatchUIDType(uid, providers.IDTypeTMDB) {
		return nil, errors.New("Not implemented")
	}

	id, err := tmdb.ParseUID(uid)
	if err != nil {
		return nil, err
	}

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

	if season == nil || show == nil || ids == nil {
		return nil, errors.New("Could not load TMDB data")
	}

	kpi, err := b.Kinopub.FindItemByIMDB(kinopub.StripImdbID(ids.ImdbID), show.OriginalName)

	if err != nil {
		return nil, err
	}

	if kpi == nil {
		return nil, errors.New("Could not find kinopub item")
	}

	if kpi, err = b.Kinopub.GetItemById(kpi.ID); kpi != nil {
		return &domain.Season{
			UID:        tmdb.ToUID(season.ID),
			Name:       season.Name,
			AirDate:    season.AirDate,
			Number:     season.SeasonNumber,
			PosterPath: season.PosterPath,
			Episodes:   toDomainEpisodes(season.SeasonNumber, season.Episodes, kpi),
		}, nil
	}

	return nil, err
}

func (b ContentBrowserImpl) GetShow(uid string) (*domain.Series, error) {
	if providers.MatchUIDType(uid, providers.IDTypeKinoHub) {
		id, _ := kinopub.ParseUID(uid)

		item, err := b.Kinopub.GetItemById(id)
		if err != nil {
			return nil, err
		}

		show, err := b.TMDB.FindTVShowByExternalID(item.ImdbID())

		if err != nil {
			return nil, err
		}

		if show != nil {
			show, err = b.TMDB.GetTVShowByID(show.ID)
			if err != nil {
				return nil, err
			}

			if show != nil {
				return show.ToDomain(), nil
			}
		}

		return item.ToDomain(), nil
	}

	if providers.MatchUIDType(uid, providers.IDTypeTMDB) {
		id, _ := tmdb.ParseUID(uid)
		show, err := b.TMDB.GetTVShowByID(id)

		if err != nil {
			return nil, err
		}

		return show.ToDomain(), nil
	}

	return nil, errors.New("Invalid UID")
}

func toDomainEpisodes(seasonNumber int, episodes []tmdb.TVEpisode, kpi *kinopub.Item) []domain.Episode {
	r := make([]domain.Episode, 0)

	for _, episode := range episodes {
		de := episode.ToDomain()

		if kpi != nil {

			if len(kpi.Seasons) >= seasonNumber {
				kps := kpi.Seasons[seasonNumber-1]

				if len(kps.Episodes) >= episode.EpisodeNumber {
					de.Files = kinopub.ToDomainFiles(kps.Episodes[episode.EpisodeNumber-1].Files)
				}
			}
		}

		r = append(r, de)
	}

	return r
}

func NewContentBrowser(kpc kinopub.KinoPubClient, tmdb tmdb.Client) ContentBrowser {
	return ContentBrowserImpl{
		Kinopub: kpc,
		TMDB:    tmdb,
	}
}
