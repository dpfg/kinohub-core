package services

import (
	"net/http"
	"strconv"

	"github.com/dpfg/kinohub-core/domain"
	httpu "github.com/dpfg/kinohub-core/pkg/http"
	provider "github.com/dpfg/kinohub-core/provider"
	"github.com/dpfg/kinohub-core/provider/kinopub"
	"github.com/dpfg/kinohub-core/provider/tmdb"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContentBrowser provides a way to find available media streams
type ContentBrowser interface {
	Show(uid string) (*domain.Series, error)
	Season(uid string, seasonNum int) (*domain.Season, error)
	Movie(uid string) (*domain.Movie, error)

	Handler() func(r chi.Router)
}

type ContentBrowserImpl struct {
	Logger  *logrus.Entry
	Kinopub kinopub.KinoPubClient
	TMDB    tmdb.Client
}

func (browser ContentBrowserImpl) Handler() func(r chi.Router) {

	return func(router chi.Router) {

		router.Get("/api/series/{series-id}", func(w http.ResponseWriter, req *http.Request) {
			uid := chi.URLParam(req, "series-id")
			show, err := browser.Show(uid)

			if err != nil {
				httpu.BadGateway(w, req, err)
				return
			}
			render.JSON(w, req, show)
		})

		router.Get("/api/series/{series-id}/seasons/{season-num}", func(w http.ResponseWriter, req *http.Request) {
			uid := chi.URLParam(req, "series-id")

			seasonNum, err := strconv.Atoi(chi.URLParam(req, "season-num"))
			if err != nil {
				httpu.BadRequest(w, req, err)
				return
			}

			season, err := browser.Season(uid, seasonNum)
			if err != nil {
				httpu.BadRequest(w, req, err)
				return
			}

			render.JSON(w, req, season)
		})

		router.Get("/api/movies/{movie-id}", func(w http.ResponseWriter, req *http.Request) {
			uid := chi.URLParam(req, "movie-id")
			m, err := browser.Movie(uid)
			if err != nil {
				httpu.BadGateway(w, req, err)
				return
			}

			render.JSON(w, req, m)
		})
	}
}

func (b ContentBrowserImpl) Season(uid string, seasonNum int) (*domain.Season, error) {
	if !provider.MatchUIDType(uid, provider.IDTypeTMDB) {
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

func (b ContentBrowserImpl) Show(uid string) (*domain.Series, error) {
	if provider.MatchUIDType(uid, provider.IDTypeKinoHub) {
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

	if provider.MatchUIDType(uid, provider.IDTypeTMDB) {
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

func (b ContentBrowserImpl) Movie(uid string) (*domain.Movie, error) {

	var imdbID string

	if provider.MatchUIDType(uid, provider.IDTypeKinoHub) {
		id, _ := kinopub.ParseUID(uid)

		item, err := b.Kinopub.GetItemById(id)
		if err != nil {
			return nil, err
		}

		imdbID = item.ImdbID()
	}

	movie, err := b.TMDB.FindMovieByExternalID(imdbID)

	if err != nil {
		return nil, err
	}

	if movie != nil {
		movie, err = b.TMDB.Movie(movie.ID)
		if err != nil {
			return nil, err
		}

		if movie != nil {
			return movie.ToDomain(), nil
		}
	}

	return nil, errors.New("Not supported UID type")
}

func NewContentBrowser(kpc kinopub.KinoPubClient, tmdb tmdb.Client) ContentBrowser {
	return ContentBrowserImpl{
		Kinopub: kpc,
		TMDB:    tmdb,
	}
}
