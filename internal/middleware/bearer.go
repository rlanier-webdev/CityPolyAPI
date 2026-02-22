package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/CityPolyAPI/internal/models"
	"gorm.io/gorm"
)

// BearerAuth validates Authorization: Bearer <token> headers
func BearerAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		rawToken := parts[1]
		sum := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(sum[:])

		var token models.AuthToken
		err := db.Where("token_hash = ? AND is_active = ?", tokenHash, true).First(&token).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive token"})
			return
		}

		// check expiration
		if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		if err := db.Model(&token).Update("last_used_at", time.Now()).Error; err != nil {
			log.Printf("bearer: failed to update last_used_at: %v", err)
		}

		// Attach userID to context for handlers
		c.Set("userID", token.UserID)
		c.Next()
	}
}