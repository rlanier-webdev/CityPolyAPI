package main

import (
	"log"
	"os"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/rlanier-webdev/CityPolyAPI/frontend"
	"github.com/rlanier-webdev/CityPolyAPI/internal/handlers"
	"github.com/rlanier-webdev/CityPolyAPI/internal/middleware"
	"github.com/rlanier-webdev/CityPolyAPI/internal/models"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	err  error
	once sync.Once
)

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

		err = db.AutoMigrate(
			&models.Game{},
			&models.User{},
			&models.AuthToken{},
			&models.APIKey{},
		)
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

	h := &handlers.Handler{DB: db}

	// Release mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Trust Railway's proxy headers
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// CORS configuration for API access
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-API-Key", "Authorization"},
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	// Rate limiting (10 req/s per IP, burst of 20)
	r.Use(middleware.RateLimitMiddleware())

	r.Static("/static", "./frontend/static")
	r.LoadHTMLGlob("frontend/templates/*")

	r.GET("/", frontend.IndexPageHandler)
	r.GET("/search", frontend.SearchPageHandler)
	r.GET("/docs", frontend.DocumentationPageHandler)
	r.GET("/games", frontend.GamesPageHandler)

	// Public auth (no middleware)
	auth := r.Group("/api/auth")
	auth.POST("/register", h.RegisterHandler)
	auth.POST("/login", h.LoginHandler)

	// Bearer-protected auth
	authBearer := auth.Group("/", middleware.BearerAuth(db))
	authBearer.POST("/logout", h.LogoutHandler)
	authBearer.POST("/keys", h.CreateAPIKeyHandler)
	authBearer.GET("/keys", h.ListAPIKeyHandler)
	authBearer.DELETE("/keys/:id", h.RevokeAPIKeyHandler)

	// API key protected data routes
	v2 := r.Group("/api/v2", middleware.APIKeyAuth(db))
	v2.GET("/games", h.GetGamesHandler)
	v2.GET("/games/:id", h.GetGameByIDHandler)
	v2.GET("/games/year/:year", h.GetGamesByYearHandler)
	v2.GET("/games/home/:team", h.GetGamesByHomeHandler)
	v2.GET("/games/away/:team", h.GetGamesByAwayHandler)
	v2.GET("/teams", h.GetTeamsHandler)

	// Public health check
	r.GET("/api", h.GetMainHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}