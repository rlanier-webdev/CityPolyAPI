package frontend

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/CityPolyAPI/models"
)

func GamesPageHandler(c *gin.Context) {
	var games []models.Game
	if err := db.Find(&games).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "games.html", gin.H{
			"Title":   "All Games",
			"Message": "Error fetching games",
		})
		return
	}

	c.HTML(http.StatusOK, "games.html", gin.H{
		"Title": "All Games",
		"Games": games,
	})
}
