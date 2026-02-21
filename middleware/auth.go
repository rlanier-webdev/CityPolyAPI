package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/RivalryAPIv2/models"
	"gorm.io/gorm"
)

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-API-Key")
		if raw == "" || !strings.HasPrefix(raw, "riv_") || len(raw) != 68 {
			c.AbortWithStatusJSON(401, gin.H{"error": "API key required"})
			return
		}
		sum := sha256.Sum256([]byte(raw))
		hash := hex.EncodeToString(sum[:])

		var key models.APIKey
		if err := db.Where("key_hash = ? AND is_active = ?", hash, true).First(&key).Error; err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or inactive API key"})
			return
		}
		// Update last_used_at without blocking the request
		go db.Model(&key).Update("last_used_at", time.Now())
		c.Set("userID", key.UserID)
		c.Next()
	}
}