package main

import (
	"log"
	"mysql/config"
	"mysql/model"
	"mysql/routes"
	"mysql/utils"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	config.LoadEnv()
	config.ConnectDatabase()

	go func() {
		for {
			time.Sleep(24 * time.Hour)
			result := config.DB.Where("expires_at < ? ", time.Now()).
				Delete(&model.Session{})
			log.Printf("Session cleanup: removed %d expired/revoked sessions", result.RowsAffected)
		}
	}()

	// Create Gin router
	r := gin.Default()

	// Apply CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // your frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-api-key", "X-Admin-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(utils.SecurityHeaders())
	// Set up routes
	routes.SetupRoutes(r)

	// Start server
	if err := r.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}
