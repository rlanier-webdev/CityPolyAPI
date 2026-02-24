package frontend

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GamesPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "games.html", gin.H{
		"Title": "All Games",
		"Games": games,
	})
}
