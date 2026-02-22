package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "SAMEORIGIN")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", "default-src 'none'; style-src 'self' cdn.jsdelivr.net fonts.googleapis.com 'unsafe-inline'; script-src cdn.jsdelivr.net kit.fontawesome.com www.googletagmanager.com 'unsafe-inline'; font-src kit.fontawesome.com fonts.gstatic.com ka-f.fontawesome.com; connect-src www.google-analytics.com ka-f.fontawesome.com cdn.jsdelivr.net; img-src 'self' data:")
        c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
        c.Next()
	}
}