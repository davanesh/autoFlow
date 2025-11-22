package main

import (
	"fmt"
	"net/http"

	"github.com/Davanesh/auto-orchestrator/internal/api"
	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()

	r := gin.Default()

	// CORS config — allow frontend at localhost:5173
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// Root test route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "AutoFlow.AI Orchestrator is running",
		})
	})

	// Register workflow routes
	api.RegisterWorkflowRoutes(r)

	// Start server
	r.Run(":8080")

	// for debug
	for _, ri := range r.Routes() {
		fmt.Println(ri.Method, ri.Path)
	}
}
