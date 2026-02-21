package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/RivalryAPIv2/helpers"
	"github.com/rlanier-webdev/RivalryAPIv2/models"
	"github.com/rlanier-webdev/RivalryAPIv2/utils"
	"golang.org/x/crypto/bcrypt"
)


func registerHandler(c *gin.Context) {
	var request models.RegisterRequest

	// Bind and validate the JSON request body
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(request.Email))

	// Validate password
	if err := utils.ValidatePassword(request.Password); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Password doesn't meet requirements."})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user model instance
	user := models.User {
		Email: email,
		PasswordHash: string(hashedPassword),
	}

	// Save user to database
	if result := db.Create(&user); result.Error != nil {
		// Handle potential unique email constraint violation or other DB errors
        c.JSON(http.StatusConflict, gin.H{"error": "Email already exists."})
		return
	}

	// Return success created
	c.JSON(http.StatusCreated, gin.H{
		"message":"User registered successfully", 
		"userID": user.ID,
	})
}

func loginHandler(c *gin.Context) {
	var request models.LoginRequest

	// Bind and validate the JSON request body
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(request.Email))

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"} )
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	rawToken, tokenHash, err := helpers.BearerToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	if err := db.Create(&models.AuthToken{
		UserID: user.ID, 
		TokenHash: tokenHash, 
		IsActive: true,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	// Return success created
	c.JSON(http.StatusOK, gin.H{
		"token": rawToken, 
	})
}

func createAPIKeyHandler(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
        return
    }

	rawKey := "riv_" + hex.EncodeToString(buf) // "riv_" + 64 hex chars = 68 total
	sum := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(sum[:])
	prefix := rawKey[4:12] // first 8 chars of hex portion

	if err := db.Create(&models.APIKey{UserID: userID, KeyHash: hash, Prefix: prefix}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H {
		"key": rawKey,
		"note": "Store this safely. It will not be shown again.",
	})
}

func listAPIKeyHandler(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var keys []models.APIKey
	if err := db.Where("user_id = ? AND is_active = ?", userID, true).Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve key"})
		return
	}

	c.JSON(http.StatusOK, keys)
}
func revokeAPIKeyHandler(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)
	id := c.Param("id")
	
	result := db.Model(&models.APIKey{}).Where("id = ? AND user_id = ?", id, userID).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to revoke key"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error":"Key doesn't exist"})
		return
	}

	c.Status(http.StatusNoContent)
}

func logoutHandler(c *gin.Context) {
	// Get the raw token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	rawToken := strings.TrimPrefix(authHeader, "Bearer ")
	rawToken = strings.TrimSpace(rawToken)
	
	// Hash the token to match what’s stored in the database
	sum := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(sum[:])

	// Look up the active token
	if err := db.Model(&models.AuthToken{}).Where("token_hash = ?", tokenHash).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.Status(http.StatusNoContent)
}