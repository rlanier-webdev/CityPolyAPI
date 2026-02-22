package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/CityPolyAPI/models"

	"gorm.io/gorm"
)
func getMainHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
		"endpoints": gin.H{
			"games": "/api/games",
			"teams": "/api/teams",
			"docs":  "/docs",
		},
	})
}

// Game Handlers
func getGamesHandler(c *gin.Context) {
	var games []models.Game
	if err := db.Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, games)
}

func getGameByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var game models.Game
	if err := db.First(&game, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, game)
}

func getGamesByYearHandler(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)

	var games []models.Game
	if err := db.Where("date >= ? AND date < ?", startDate, endDate).Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, games)
}

func getGamesByHomeHandler(c *gin.Context) {
	homeTeam := c.Param("team")

	if homeTeam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team name cannot be empty"})
		return
	}

	var games []models.Game
	if err := db.Where("home_team = ?", homeTeam).Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrive games: " + err.Error()})
		return
	}

	if len(games) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No games found for the given team"})
		return
	}

	c.JSON(http.StatusOK, games)
}

func getGamesByAwayHandler(c *gin.Context) {
	homeTeam := c.Param("team")

	if homeTeam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team name cannot be empty"})
		return
	}

	var games []models.Game
	if err := db.Where("away_team = ?", homeTeam).Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrive games: " + err.Error()})
		return
	}

	if len(games) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No games found for the given team"})
		return
	}

	c.JSON(http.StatusOK, games)
}

// End game handlers

// Team Handlers
func getTeamsHandler(c *gin.Context) {
	var homeTeams, awayTeams []string

	if err := db.Model(&models.Game{}).Distinct().Pluck("home_team", &homeTeams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&models.Game{}).Distinct().Pluck("away_team", &awayTeams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Combine and deduplicate
	teamSet := make(map[string]bool)
	for _, t := range homeTeams {
		teamSet[t] = true
	}
	for _, t := range awayTeams {
		teamSet[t] = true
	}

	teams := make([]string, 0, len(teamSet))
	for t := range teamSet {
		teams = append(teams, t)
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// End team handlers
