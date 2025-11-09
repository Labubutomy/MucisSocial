package pkg

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware логирует HTTP запросы для Gin
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Printf("%s %s %s %d %s",
			param.Method,
			param.Path,
			param.ClientIP,
			param.StatusCode,
			param.Latency,
		)
		return ""
	})
}

// SimpleLoggingMiddleware простое логирование для Gin
func SimpleLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %s", c.Request.Method, c.Request.RequestURI, time.Since(start))
	}
}
