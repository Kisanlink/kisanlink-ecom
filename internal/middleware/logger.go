package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger provides structured logging for HTTP requests.
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"timestamp":  param.TimeStamp.Format("2006-01-02 15:04:05"),
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"latency":    param.Latency,
			"request_id": param.Request.Header.Get("X-Request-ID"),
		}).Info("HTTP Request")

		return ""
	})
}
