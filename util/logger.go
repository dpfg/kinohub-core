package util

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HTTPLogger is the logrus logger handler
func HTTPLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		if len(c.Request.URL.RawQuery) > 0 {
			path = path + "?" + c.Request.URL.Query().Encode()
		}
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := log.WithField("prefix", "http")

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s %s [%d] (%dms)", clientIP, c.Request.Method, path, statusCode, latency)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
