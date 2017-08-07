package main

import (
	"net/http"

	"github.com/dpfg/kinohub-core/providers/kinopub"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/creds", func(c *gin.Context) {
		cl := kinopub.KinoPubClientImpl{
			ClientID:     "plex",
			ClientSecret: "h2zx6iom02t9cxydcmbo9oi0llld7jsv",
			TokenStorage: kinopub.JSONTokenStorage{
				Path: "config.json",
			},
		}

		r, err := cl.SearchItemBy(kinopub.ItemsFilter{
			Title: "trans",
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(200, r)
	})

	r.Run("0.0.0.0:8081") // listen and serve on 0.0.0.0:8080
}
