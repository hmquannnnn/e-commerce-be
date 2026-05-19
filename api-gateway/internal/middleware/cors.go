package middleware

import (
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var allowedOrigins = []string{
	"http://localhost:3000",
	"http://127.0.0.1:3000",
	"https://uav-store.io.vn",
	"https://www.uav-store.io.vn",
}

func init() {
	rawOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	for _, origin := range strings.Split(rawOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}
}

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	if slices.Contains(allowedOrigins, origin) {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme == "https" && strings.HasSuffix(host, ".vercel.app") {
		return true
	}
	return false
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if isAllowedOrigin(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID, ngrok-skip-browser-warning")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE, HEAD")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
