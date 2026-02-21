package models

import (
	"time"

	"github.com/rlanier-webdev/RivalryAPIv2/utils"
)

type Game struct {
	ID            uint `gorm:"primaryKey"`
	HomeTeam      string
	AwayTeam      string
	Date          utils.CustomDate
	HomeTeamScore int
	AwayTeamScore int
	Notes         string
}

type User struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Email        string    `gorm:"uniqueIndex;not null" json:"email"`
    PasswordHash string    `gorm:"not null" json:"-"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    APIKeys      []APIKey  `gorm:"foreignKey:UserID" json:"-"`
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type APIKey struct {
    ID         uint       `gorm:"primaryKey" json:"id"`
    UserID     uint       `gorm:"not null;index" json:"user_id"`
    KeyHash    string     `gorm:"not null;uniqueIndex" json:"-"`
    Prefix     string     `gorm:"not null" json:"prefix"`    // first 8 chars for display
    Name       string     `json:"name"`
    IsActive   bool       `gorm:"default:true" json:"is_active"`
    LastUsedAt *time.Time `json:"last_used_at"`
    CreatedAt  time.Time  `json:"created_at"`
}