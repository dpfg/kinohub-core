package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/trakt"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	r := gin.New()

	logger := logrus.StandardLogger()
	// logger.Level = logrus.DebugLevel
	r.Use(ginrus.Ginrus(logger, time.RFC3339, true))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/creds", func(c *gin.Context) {
		cl := kinopub.KinoPubClientImpl{
			ClientID:     "plex",
			ClientSecret: "h2zx6iom02t9cxydcmbo9oi0llld7jsv",
			PreferenceStorage: providers.JSONPreferenceStorage{
				Path: ".data/",
			},
		}

		r, err := cl.SearchItemBy(kinopub.ItemsFilter{
			Title: "game of th",
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(200, r)
	})

	r.GET("/trakt/signin", func(c *gin.Context) {
		cl := trakt.NewTraktClient()
		// f0429b45753645dae219dcf44d673e4eda082dd1dc0f808e925c5e78b6184019
		c.JSON(http.StatusOK, cl.GetAuthCodeURL())
	})

	r.GET("/trakt/exchange", func(c *gin.Context) {
		cl := trakt.NewTraktClient()
		t, err := cl.Exchange(context.Background(), "f0429b45753645dae219dcf44d673e4eda082dd1dc0f808e925c5e78b6184019")
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, t)
	})

	r.GET("/trakt/shows/tranding", func(c *gin.Context) {
		cl := trakt.NewTraktClient()
		m, err := cl.GetMyShows(12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		logger.Debugln("Show log here")
		kpc := kinopub.KinoPubClientImpl{
			ClientID:     "plex",
			ClientSecret: "h2zx6iom02t9cxydcmbo9oi0llld7jsv",
			PreferenceStorage: providers.JSONPreferenceStorage{
				Path: ".data/",
			},
		}

		r := make([]interface{}, 0)
		for _, item := range m {
			id, _ := strconv.Atoi(strings.TrimLeft(item.Show.Ids.Imdb, "tt"))
			log.Println(item.Episode)
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

	r.Run("0.0.0.0:8081") // listen and serve on 0.0.0.0:8080
}
