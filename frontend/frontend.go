package frontend

import (
	"github.com/rlanier-webdev/CityPolyAPI/models"
	"gorm.io/gorm"
)

var db *gorm.DB

var games []models.Game

func SetDB(database *gorm.DB) {
	db = database
}

func SetGames(g []models.Game) {
	games = g
}