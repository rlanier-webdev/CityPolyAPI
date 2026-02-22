package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/CityPolyAPI/internal/helpers"
	"github.com/rlanier-webdev/CityPolyAPI/internal/models"
	"github.com/rlanier-webdev/CityPolyAPI/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type loginAttempt struct {
	count       int
	lockedUntil time.Time
	lastFailed  time.Time
}

var (
	loginAttempts   = make(map[string]*loginAttempt)
	loginAttemptsMu sync.Mutex
)

const (
	maxLoginFailures   = 10
	loginLockoutPeriod = 15 * time.Minute
)

func init() {
	go cleanupLoginAttempts()
}

func cleanupLoginAttempts() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		loginAttemptsMu.Lock()
		for email, attempt := range loginAttempts {
			if time.Since(attempt.lastFailed) > 30*time.Minute {
				delete(loginAttempts, email)
			}
		}
		loginAttemptsMu.Unlock()
	}
}


func (h *Handler) RegisterHandler(c *gin.Context) {
	var request models.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("RegisterHandler: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))

	if err := utils.ValidatePassword(request.Password); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Password doesn't meet requirements."})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if result := h.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *Handler) LoginHandler(c *gin.Context) {
	var request models.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("LoginHandler: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	
	loginAttemptsMu.Lock()
	attempt, exists := loginAttempts[email]
	if exists && time.Now().Before(attempt.lockedUntil) {
		loginAttemptsMu.Unlock()
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Account temporarily locked. Try again later."})
		return
	}
	loginAttemptsMu.Unlock()

	var user models.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		loginAttemptsMu.Lock()
		if loginAttempts[email] == nil {
			loginAttempts[email] = &loginAttempt{}
		}
		loginAttempts[email].count++
		loginAttempts[email].lastFailed = time.Now()
		if loginAttempts[email].count >= maxLoginFailures {
			loginAttempts[email].lockedUntil = time.Now().Add(loginLockoutPeriod)
			loginAttempts[email].count = 0
		}
		loginAttemptsMu.Unlock()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	loginAttemptsMu.Lock()
	delete(loginAttempts, email)
	loginAttemptsMu.Unlock()

	rawToken, tokenHash, err := helpers.BearerToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	tokenExpiry := time.Now().Add(30 * 24 * time.Hour)
	if err := h.DB.Create(&models.AuthToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		IsActive:  true,
		ExpiresAt: &tokenExpiry,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": rawToken,
	})
}

func (h *Handler) CreateAPIKeyHandler(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}

	rawKey := "riv_" + hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(sum[:])
	prefix := rawKey[4:12]

	if err := h.DB.Create(&models.APIKey{UserID: userID, KeyHash: hash, Prefix: prefix}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":  rawKey,
		"note": "Store this safely. It will not be shown again.",
	})
}

func (h *Handler) ListAPIKeyHandler(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var keys []models.APIKey
	if err := h.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve key"})
		return
	}

	c.JSON(http.StatusOK, keys)
}

func (h *Handler) RevokeAPIKeyHandler(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}
	
	result := h.DB.Model(&models.APIKey{}).Where("id = ? AND user_id = ?", id, userID).Update("is_active", false)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke key"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key doesn't exist"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) LogoutHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	rawToken := strings.TrimPrefix(authHeader, "Bearer ")
	rawToken = strings.TrimSpace(rawToken)

	sum := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(sum[:])

	if err := h.DB.Model(&models.AuthToken{}).Where("token_hash = ?", tokenHash).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.Status(http.StatusNoContent)
}