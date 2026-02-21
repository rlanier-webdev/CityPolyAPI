package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/rlanier-webdev/RivalryAPIv2/frontend"
	"github.com/rlanier-webdev/RivalryAPIv2/models"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	err      error
	once     sync.Once
	limiters = make(map[string]*rate.Limiter)
	limiterMu sync.Mutex
)

// rateLimitMiddleware limits requests per IP address
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiterMu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			// Allow 10 requests per second with burst of 20
			limiter = rate.NewLimiter(10, 20)
			limiters[ip] = limiter
		}
		limiterMu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please slow down.",
			})
			return
		}
		c.Next()
	}
}

func init() {
	// Load .env file if present (local dev only, ignored in production)
	_ = godotenv.Load()
}

func initDB() {
	once.Do(func() {
		db, err = gorm.Open(sqlite.Open("games.db"), &gorm.Config{})
		if err != nil {
			log.Fatal("failed to connect to database: ", err)
		}

		err = db.AutoMigrate(&models.Game{}, &models.User{}, &models.APIKey{})
		if err != nil {
			log.Fatal("failed to migrate database: ", err)
		}

		// Load games from the database
		var games []models.Game
		err = db.Find(&games).Error
		if err != nil {
			log.Fatal("failed to load games from the database: ", err)
		}

		// Set the loaded games in the frontend package
		frontend.SetGames(games)
	})
}

func main() {
	initDB()
	frontend.SetDB(db)

	// Release mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Trust Railway's proxy headers
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// CORS configuration for API access
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	// Rate limiting (10 req/s per IP, burst of 20)
	r.Use(rateLimitMiddleware())

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", frontend.IndexPageHandler)
	r.GET("/search", frontend.SearchPageHandler)
	r.GET("/docs", frontend.DocumentationPageHandler)
	r.GET("/games", frontend.GamesPageHandler)

	r.GET("/api", getMainHandler)
	r.GET("/api/games", getGamesHandler)
	r.GET("/api/games/:id", getGameByIDHandler)
	r.GET("/api/games/year/:year", getGamesByYearHandler)
	r.GET("/api/games/home/:team", getGamesByHomeHandler)
	r.GET("/api/games/away/:team", getGamesByAwayHandler)

	r.GET("/api/teams", getTeamsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
