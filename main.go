package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	cors "gopkg.in/gin-contrib/cors.v1"

	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/trakt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	r.Use(ginlogrus.Logger(logger))

	// Initialize common cache manager that will be used by API clients
	cacheManager, err := providers.NewCacheManager(logger)
	if err != nil {
		logrus.Errorf("Cannot initialize cache manager. %s", err.Error())
		return
	}

	kpc := kinopub.KinoPubClientImpl{
		ClientID:     "plex",
		ClientSecret: "h2zx6iom02t9cxydcmbo9oi0llld7jsv",
		PreferenceStorage: providers.JSONPreferenceStorage{
			Path: ".data/",
		},
		CacheFactory: cacheManager,
		Logger:       logger,
	}

	tc := trakt.NewTraktClient(logger)

	r.GET("/search", func(c *gin.Context) {
		r, err := kpc.SearchItemBy(kinopub.ItemsFilter{
			Title: c.Query("q"),
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(200, r)
	})

	r.GET("/items/:item-id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("item-id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		item, err := kpc.GetItemById(int64(id))
		if err != nil {
			c.JSON(http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(200, item)
	})

	r.GET("/tv/releases", func(c *gin.Context) {
		from, _ := time.Parse("2006-01-02", c.Query("from"))
		to, _ := time.Parse("2006-01-02", c.Query("to"))

		m, err := tc.GetMyShows(from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		r := make([]interface{}, 0)
		for _, item := range m {
			id, _ := strconv.Atoi(strings.TrimLeft(item.Show.Ids.Imdb, "tt"))
			logrus.Debugln("--------------------------------------")
			ep, err := kpc.GetEpisode(id, item.Show.Title, item.Episode.Season, item.Episode.Number)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}

			r = append(r, struct {
				ShowTitle   string
				EpisodeTite string
				FirstAired  time.Time
				Episode     interface{}
			}{
				ShowTitle:   item.Show.Title,
				EpisodeTite: item.Episode.Title,
				FirstAired:  item.FirstAired,
				Episode:     ep,
			})
		}

		c.JSON(http.StatusOK, r)
	})

	// r.GET("/trakt/signin", func(c *gin.Context) {
	// 	cl := trakt.NewTraktClient()
	// 	// f0429b45753645dae219dcf44d673e4eda082dd1dc0f808e925c5e78b6184019
	// 	c.JSON(http.StatusOK, cl.GetAuthCodeURL())
	// })

	// r.GET("/trakt/exchange", func(c *gin.Context) {
	// 	cl := trakt.NewTraktClient()
	// 	t, err := cl.Exchange(context.Background(), "f0429b45753645dae219dcf44d673e4eda082dd1dc0f808e925c5e78b6184019")
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, err.Error())
	// 		return
	// 	}

	// 	c.JSON(http.StatusOK, t)
	// })

	r.Run("0.0.0.0:8081") // listen and serve on 0.0.0.0:8080
}
