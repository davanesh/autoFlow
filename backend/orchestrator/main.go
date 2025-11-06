package main

import (
	"net/http"

	"github.com/Davanesh/auto-orchestrator/internal/api"
	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/gin-gonic/gin"
)

func main() {
    db.InitDB()
	r := gin.Default()
	
	r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H {
            "message": "AutoFlow.AI Orchestrator is running",
        })
    })
    api.RegisterWorkflowRoutes(r)
    // r.POST("/workflows/:id/run", api.RunWorkflow)
    r.Run(":8080")
}