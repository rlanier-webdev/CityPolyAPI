// frontend/search.go
package frontend

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rlanier-webdev/CityPolyAPI/internal/models"
)

func SearchPageHandler(c *gin.Context) {
	searchType := c.Query("searchType")
	query := c.Query("query")
	var results []models.Game
	var message string

	switch searchType {
	case "id":
		id, err := strconv.Atoi(query)
		if err == nil {
			for _, g := range games {
				if g.ID == uint(id) {
					results = append(results, g)
				}
			}
		}
	case "home":
		for _, g := range games {
			if g.HomeTeam == query {
				results = append(results, g)
			}
		}
	case "away":
		for _, g := range games {
			if g.AwayTeam == query {
				results = append(results, g)
			}
		}
	case "year":
		year, err := strconv.Atoi(query)
		if err == nil {
			for _, g := range games {
				if g.Date.Time.Year() == year {
					results = append(results, g)
				}
			}
		}
	}

	if len(results) == 0 {
		message = "No results found"
	}

	c.HTML(http.StatusOK, "search.html", gin.H{
		"Title":   "Search Results",
		"Results": results,
		"Message": message,
	})
}
