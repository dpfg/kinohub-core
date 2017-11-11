package services

import (
	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/tmdb"
	"github.com/sirupsen/logrus"
)

// ContentSearch provides a way to find available media streams
type ContentSearch interface {
	Search(q string) ([]domain.SearchResult, error)
}

type ContentSearchImpl struct {
	Logger  *logrus.Entry
	Kinopub kinopub.KinoPubClient
	TMDB    tmdb.Client
}

func (cs ContentSearchImpl) Search(q string) ([]domain.SearchResult, error) {
	r, err := cs.Kinopub.SearchItemBy(kinopub.ItemsFilter{
		Title: q,
	})

	if err != nil {
		return nil, err
	}

	result := make([]domain.SearchResult, 0)
	for _, item := range r {
		// tmdbItem, err := cs.TMDB.FindTVShowByExternalID(item.ImdbID())
		// if err != nil {
		// 	cs.Logger.Error(errors.WithMessage(err, "cannot load TMDB entry by IMDB id").Error())
		// 	continue
		// }

		// if tmdbItem == nil {
		// 	cs.Logger.Debugf("Skip search result item [%d]", item.ID)
		// 	continue
		// }

		result = append(result, domain.SearchResult{
			Type:       item.DomainType(),
			Title:      item.Title,
			PosterPath: item.Posters.Big,
			// TmdbID: i.ID,
		})
	}

	return result, nil
}
