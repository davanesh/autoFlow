package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/Davanesh/auto-orchestrator/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterWorkflowRoutes(r *gin.Engine) {
	r.GET("/workflows", GetWorkflows)
	r.POST("/workflows", CreateWorkflow)
	r.PUT("/workflows/:id", UpdateWorkflowStatus)
}

func UpdateWorkflowStatus(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Status string `json:"status"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🧩 Convert string ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"status": body.Status}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workflow status updated successfully",
		"id":      id,
		"status":  body.Status,
	})
}


func GetWorkflows(c *gin.Context) {
	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var workflows []models.Workflow
	if err := cursor.All(ctx, &workflows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflows)
}

func CreateWorkflow(c *gin.Context) {
	var wf models.Workflow
	if err := c.BindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wf.CreatedAt = time.Now()
	wf.Status = "draft"

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, wf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	wf.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, wf)
}
