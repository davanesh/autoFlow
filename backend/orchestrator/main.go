package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	
	_ "github.com/Davanesh/auto-orchestrator/internal/executors"
	"github.com/Davanesh/auto-orchestrator/internal/api"
	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// -------------------------------
	// 1) Load Environment Variables
	// -------------------------------
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env file not found, using system env")
	}
	
	// Optional: check if OPENAI_KEY is loaded
	log.Println("🔑 OPENAI_KEY Loaded:", os.Getenv("OPENAI_API_KEY") != "")

	// -------------------------------
	// 2) Initialize Database
	// -------------------------------
	db.InitDB()

	// -------------------------------
	// 3) Setup Gin
	// -------------------------------
	r := gin.Default()

	// CORS setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// Root route for testing
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "AutoFlow.AI Orchestrator is running",
		})
	})

	// Register workflow routes
	api.RegisterWorkflowRoutes(r)

	// -------------------------------
	// 4) Start Server
	// -------------------------------
	go func() {
		r.Run(":8080")
	}()

	// -------------------------------
	// 5) Print all registered routes
	// -------------------------------
	for _, route := range r.Routes() {
		fmt.Println(route.Method, route.Path)
	}

	// Prevent main from exiting
	select {}
}
