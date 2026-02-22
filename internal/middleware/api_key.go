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

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("X-API-Key")
	
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"Authorization header required"})
			return
		}

		// Validate format
		if !strings.HasPrefix(authHeader, "riv_") || len(authHeader) != 68 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"Authorization header required"})
			return
		}

		sum := sha256.Sum256([]byte(authHeader))
		keyHash := hex.EncodeToString(sum[:])

		var key models.APIKey
		err := db.Where("key_hash = ? AND is_active = ?", keyHash, true).First(&key).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive key"})
			return
		}

		if err := db.Model(&key).Update("last_used_at", time.Now()).Error; err != nil {
		log.Printf("api_key: failed to update last_used_at: %v", err)
	}

		c.Set("userID", key.UserID)
		c.Next()

	}
	
}