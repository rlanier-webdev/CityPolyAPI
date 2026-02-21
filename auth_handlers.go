package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/RivalryAPIv2/models"
	"golang.org/x/crypto/bcrypt"
)

func validatePassword(password string) error {
	var errs []string

	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if !hasUpper {
		errs = append(errs, "must contain at least one uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "must contain at least one lowercase letter")
	}
	if !hasNumber {
		errs = append(errs, "must contain at least one number")
	}
	if !hasSymbol {
		errs = append(errs, "must contain at least one symbol (e.g. !@#$%)")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func bearerToken() (rawToken string, tokenHash string, err error) {
	prefix := "bearer_"

	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Hex encode
	encoded := hex.EncodeToString(bytes)

	// Add prefix
	token := prefix + encoded

	// Create a new hash instance
	hasher := sha256.New()

	// Write the token bytes to the hasher
	hasher.Write([]byte(token))

	// Get the finalized hash result as a byte slice
	tokenHash = hex.EncodeToString(hasher.Sum(nil))

	// Encode the byte slice into a human-readable hexadecimal string
	return token, tokenHash, nil
}

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
	if err := validatePassword(request.Password); err != nil {
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

	rawToken, tokenHash, err := bearerToken()
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
	// Extract from header
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"Header Required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid format"}) //
		return
	}

	decoded := base64.StdEncoding.DecodeString(parts[1])
	email := 
	password := strings.SplitN(string(decoded), ":", 2)

	buf := make([]byte, 32)
	rand.Read(buf)                             // crypto/rand
	rawKey := "riv_" + hex.EncodeToString(buf) // "riv_" + 64 hex chars = 68 total
	sum := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(sum[:])
	prefix := rawKey[4:12] // first 8 chars of hex portion
}
func listAPIKeyHandler(c *gin.Context) {

}
func revokeAPIKeyHandler(c *gin.Context) {

}