package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Davanesh/auto-orchestrator/internal/api"
	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/Davanesh/auto-orchestrator/internal/executors" // IMPORTANT: kept for webhook handler
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

	log.Println("🔑 OPENAI_KEY Loaded:", os.Getenv("OPENAI_API_KEY") != "")
	log.Println("🔑 TWILIO SID Loaded:", os.Getenv("TWILIO_SID") != "")
	log.Println("🔑 ALLOWED_WHATSAPP_NUMBER:", os.Getenv("ALLOWED_WHATSAPP_NUMBER"))

	// -------------------------------
	// 2) Initialize Database
	// -------------------------------
	db.InitDB()

	// -------------------------------
	// 3) Setup Gin Server
	// -------------------------------
	r := gin.Default()

	// CORS setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// -------------------------------
	// 4) Basic Health Route
	// -------------------------------
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "AutoFlow.AI Orchestrator is running",
		})
	})

	// -------------------------------
	// 5) Workflow Routes
	// -------------------------------
	api.RegisterWorkflowRoutes(r)

	// -------------------------------
	// 6) WhatsApp Webhook Route
	// -------------------------------
	r.POST("/webhook/whatsapp", func(c *gin.Context) {
		// Use our executor handler (converted to Gin)
		executors.HandleWhatsAppWebhookGin(c)
	})

	// -------------------------------
	// 7) Start the Server (async)
	// -------------------------------
	go func() {
		log.Println("🚀 Orchestrator running on port 8080...")
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Server failed:", err)
		}
	}()

	// -------------------------------
	// 8) Print all Registered Routes
	// -------------------------------
	for _, route := range r.Routes() {
		fmt.Println(route.Method, route.Path)
	}

	// Keep process alive
	select {}
}
