package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	cors "gopkg.in/gin-contrib/cors.v1"

	"github.com/dpfg/kinohub-core/providers"
	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/dpfg/kinohub-core/providers/tmdb"
	"github.com/dpfg/kinohub-core/providers/trakt"
	"github.com/dpfg/kinohub-core/services"
	"github.com/dpfg/kinohub-core/util"
	"github.com/gin-gonic/gin"
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const (
	defaultPort = 8081

	zeroConfName    = "KinoHub"
	zeroConfService = "_kinohub._tcp"
	zeroConfDomain  = "local."
)

func main() {

	// Setup logger
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)

	logfmt := new(prefixed.TextFormatter)
	// logfmt.DisableColors = true
	logfmt.FullTimestamp = true
	logfmt.TimestampFormat = "2006/01/02 15:04:05"
	logger.Formatter = logfmt

	// Register as a zero config service
	logger.Infof("Starting zeroconf service [%s]\n", zeroConfName)
	server, err := zeroconf.Register(zeroConfName, zeroConfService, zeroConfDomain, defaultPort, nil, nil)
	if err != nil {
		logger.Errorf("Cannot start zeroconf service: %s\n", err.Error())
	}
	defer server.Shutdown()

	router := gin.New()
	router.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	router.Use(util.HTTPLogger(logger))

	// Initialize common cache manager that will be used by API clients
	cacheFactory, err := providers.NewCacheFactory(logger)
	if err != nil {
		logrus.Errorf("Cannot initialize cache factory. %s", err.Error())
		return
	}

	// Initialize common preference storage
	ps := providers.JSONPreferenceStorage{
		Path: ".data/",
	}

	kpc := kinopub.NewKinoPubClient(logger, cacheFactory)
	tc := trakt.NewTraktClient(logger)
	feed := services.NewFeed(tc, kpc, logger)
	tmdbc := tmdb.New(logger, cacheFactory, ps)
	browser := services.NewContentBrowser(kpc, tmdbc)

	router.GET("/series/:series-id", func(c *gin.Context) {
		uid := c.Param("series-id")
		show, err := browser.GetShow(uid)

		if err != nil {
			httpError(c, http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(http.StatusOK, show)
	})

	router.GET("/series/:series-id/seasons/:season-num", func(c *gin.Context) {
		uid := c.Param("series-id")

		seasonNum, err := strconv.Atoi(c.Param("season-num"))
		if err != nil {
			httpError(c, http.StatusBadRequest, err.Error())
			return
		}

		season, err := browser.GetSeason(uid, seasonNum)
		if err != nil {
			httpError(c, http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(http.StatusOK, season)
	})

	router.GET("/search", func(c *gin.Context) {
		r, err := kpc.SearchItemBy(kinopub.ItemsFilter{
			Title: c.Query("q"),
		})

		if err != nil {
			httpError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(200, r)
	})

	router.GET("/search2", func(c *gin.Context) {
		search := services.ContentSearchImpl{Kinopub: kpc, TMDB: tmdbc, Logger: logger.WithField("prefix", "search")}
		result, err := search.Search(c.Query("q"))
		if err != nil {
			httpError(c, http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(200, result)
	})

	router.GET("/items/:item-id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("item-id"))
		if err != nil {
			httpError(c, http.StatusBadRequest, err.Error())
			return
		}

		item, err := kpc.GetItemById(id)
		if err != nil {
			httpError(c, http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(200, item)
	})

	router.GET("/tv/releases", func(c *gin.Context) {
		from, _ := time.Parse("2006-01-02", c.Query("from"))
		to, _ := time.Parse("2006-01-02", c.Query("to"))

		releases, err := feed.Releases(from, to)
		if err != nil {
			httpError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, releases)
	})

	router.POST("/scrobble/:tmdbc-id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("tmdbc-id"))
		if err != nil {
			httpError(c, http.StatusBadRequest, err.Error())
			return
		}

		err = tc.Scrobble(id)
		if err != nil {
			httpError(c, http.StatusBadGateway, "Cannot scrobble item: "+err.Error())
			return
		}
	})

	router.GET("/trakt/trending", func(c *gin.Context) {
		shows, err := tc.GetTrendingShows()
		if err != nil {
			httpError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, shows)
	})

	router.GET("/trakt/signin", func(c *gin.Context) {
		c.JSON(http.StatusOK, tc.GetAuthCodeURL())
	})

	router.GET("/trakt/exchange", func(c *gin.Context) {
		t, err := tc.Exchange(context.Background(), c.Param("code"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, t)
	})

	router.Run(fmt.Sprintf("0.0.0.0:%d", defaultPort)) // listen and serve on 0.0.0.0:8080
}

func httpError(c *gin.Context, code int, msg string) {
	c.JSON(code, struct{ Msg string }{Msg: msg})
}
