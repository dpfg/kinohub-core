package services

import (
	"net/http"

	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/provider/kinopub"
	"github.com/dpfg/kinohub-core/provider/tmdb"
	"github.com/dpfg/kinohub-core/util"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

// ContentSearch provides a way to find available media streams
type ContentSearch struct {
	Logger  *logrus.Entry
	Kinopub kinopub.KinoPubClient
	TMDB    tmdb.Client
}

// Handler return http.Handler that can servce search-related requests
func (cs ContentSearch) Handler() http.Handler {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, req *http.Request) {
		result, err := cs.Search(req.URL.Query().Get("q"))
		if err != nil {
			util.InternalError(w, req, err)
			return
		}
		render.JSON(w, req, result)
	})

	return router
}

func (cs ContentSearch) Search(q string) ([]domain.SearchResult, error) {
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
			UID:        kinopub.ToUID(item.ID),
			Type:       item.DomainType(),
			Title:      item.Title,
			PosterPath: item.Posters.Big,
		})
	}

	return result, nil
}
