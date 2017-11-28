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

	if err != nil {
		return nil, err
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
	if providers.MatchUIDType(uid, providers.ID_TYPE_KINOHUB) {
		id, _ := kinopub.ParseUID(uid)

		item, err := b.Kinopub.GetItemById(id)
		if err != nil {
			return nil, err
		}

		show, err := b.TMDB.FindTVShowByExternalID(item.ImdbID())

		if err != nil {
			return nil, err
		}

		show, err = b.TMDB.GetTVShowByID(show.ID)
		if err != nil {
			return nil, err
		}

		return show.ToDomain(), nil
	}

	if providers.MatchUIDType(uid, providers.ID_TYPE_TMDB) {
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
		de := domain.Episode{
			Number: episode.EpisodeNumber,
			// FirstAired: episode.AirDate, // TODO:
			Overview:  episode.Overview,
			Title:     episode.Name,
			UID:       tmdb.ToUID(episode.ID),
			StillPath: tmdb.ImagePath(episode.StillPath, tmdb.OriginalSize),
		}

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
