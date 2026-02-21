package main

import (
	"crypto/rand"
	"crypto/sha256"
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
	var request models.RegisterRequest

	// Bind and validate the JSON request body
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"} )
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// Return success created
	c.JSON(http.StatusCreated, gin.H{
		"message":"Login successfully", 
		"userID": user.ID,
	})
}

func createAPIKeyHandler(c *gin.Context) {
	buf := make([]byte, 32)
	rand.Read(buf)                             // crypto/rand
	rawKey := "riv_" + hex.EncodeToString(buf) // "riv_" + 64 hex chars = 68 total
	sum := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(sum[:])
	prefix := rawKey[4:12] // first 8 chars of hex portion
}
func listAPIKeysHandler(c *gin.Context) {

}
func revokeAPIKeyHandler(c *gin.Context) {

}