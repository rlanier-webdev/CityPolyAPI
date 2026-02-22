package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/CityPolyAPI/internal/helpers"
	"github.com/rlanier-webdev/CityPolyAPI/internal/models"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func (h *Handler) GetMainHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
		"endpoints": gin.H{
			"games": "/api/v2/games",
			"teams": "/api/v2/teams",
			"docs":  "/docs",
		},
	})
}

// Game Handlers
func (h *Handler) GetGamesHandler(c *gin.Context) {
	limit, offset := helpers.ParsePagination(c)
	var games []models.Game
	if err := h.DB.Limit(limit).Offset(offset).Find(&games).Error; err != nil {
		log.Printf("GetGamesHandler: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, games)
}

func (h *Handler) GetGameByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var game models.Game
	if err := h.DB.First(&game, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		} else {
			log.Printf("GetGameByIDHandler: db error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, game)
}

func (h *Handler) GetGamesByYearHandler(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)

	var games []models.Game
	if err := h.DB.Where("date >= ? AND date < ?", startDate, endDate).Find(&games).Error; err != nil {
		log.Printf("GetGamesByYearHandler: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, games)
}

func (h *Handler) GetGamesByHomeHandler(c *gin.Context) {
	homeTeam := c.Param("team")

	if homeTeam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team name cannot be empty"})
		return
	}

	var games []models.Game
	if err := h.DB.Where("home_team = ?", homeTeam).Find(&games).Error; err != nil {
		log.Printf("GetGamesByHomeHandler: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if len(games) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No games found for the given team"})
		return
	}

	c.JSON(http.StatusOK, games)
}

func (h *Handler) GetGamesByAwayHandler(c *gin.Context) {
	awayTeam := c.Param("team")

	if awayTeam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team name cannot be empty"})
		return
	}

	var games []models.Game
	if err := h.DB.Where("away_team = ?", awayTeam).Find(&games).Error; err != nil {
		log.Printf("GetGamesByAwayHandler: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if len(games) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No games found for the given team"})
		return
	}

	c.JSON(http.StatusOK, games)
}

// Team Handlers
func (h *Handler) GetTeamsHandler(c *gin.Context) {
	var homeTeams, awayTeams []string

	if err := h.DB.Model(&models.Game{}).Distinct().Pluck("home_team", &homeTeams).Error; err != nil {
		log.Printf("GetTeamsHandler: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if err := h.DB.Model(&models.Game{}).Distinct().Pluck("away_team", &awayTeams).Error; err != nil {
		log.Printf("GetTeamsHandler: db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
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
