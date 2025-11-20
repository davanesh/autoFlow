package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/Davanesh/auto-orchestrator/internal/models"
	"github.com/Davanesh/auto-orchestrator/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---------------- ROUTE REGISTRATION ----------------

func RegisterWorkflowRoutes(r *gin.Engine) {
	r.GET("/workflows", GetWorkflows)
	r.POST("/workflows", CreateWorkflow)
	r.PUT("/workflows/:id", UpdateWorkflowStatus)
	r.POST("/workflows/:id/run", RunWorkflow)
	r.PUT("/workflows/:id/structure", SaveWorkflowStructure)
}

// ---------------- UPDATE STATUS ----------------

func UpdateWorkflowStatus(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Status string `json:"status"`
	}

	if err := c.BindJSON(&body); err != nil {
		log.Println("❌ Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"status": body.Status}}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated",
		"status":  body.Status,
	})
}

// ---------------- GET ALL WORKFLOWS ----------------

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

// ---------------- CREATE WORKFLOW ----------------

func CreateWorkflow(c *gin.Context) {
	var wf models.Workflow

	if err := c.BindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow JSON"})
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

// ---------------- RUN WORKFLOW (REAL LAMBDA INVOKE) ----------------

func RunWorkflow(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wf models.Workflow
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&wf); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Lambda client
	lambdaInvoker, err := services.NewLambdaInvoker()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize AWS Lambda client"})
		return
	}

	logCollection := db.GetCollection("execution_logs")
	results := []models.ExecutionLog{}

	// Execute the workflow sequentially
	for i, node := range wf.Nodes {

		if node.LambdaName == "" {
			log.Println("⚠️ Node missing lambdaName field:", node.CanvasID)
			continue
		}

		taskName := node.LambdaName
		log.Printf("🚀 Invoking Lambda: %s\n", taskName)

		// Create "running" log
		runningLog := models.ExecutionLog{
			WorkflowID:  wf.ID.Hex(),
			TaskName:    taskName,
			Status:      "running",
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Lambda %s started", taskName),
		}
		logCollection.InsertOne(ctx, runningLog)

		// Payload sent to actual Lambda
		payload := map[string]interface{}{
			"workflowId": wf.ID.Hex(),
			"node":       node,
			"time":       time.Now().Unix(),
		}

		respPayload, invokeErr := lambdaInvoker.Invoke(taskName, payload)

		// Log result
		var execLog models.ExecutionLog

		if invokeErr != nil {
			execLog = models.ExecutionLog{
				WorkflowID:  wf.ID.Hex(),
				TaskName:    taskName,
				Status:      "failed",
				Timestamp:   time.Now(),
				Description: invokeErr.Error(),
			}
		} else {
			execLog = models.ExecutionLog{
				WorkflowID:  wf.ID.Hex(),
				TaskName:    taskName,
				Status:      "completed",
				Timestamp:   time.Now(),
				Description: string(respPayload),
			}
		}

		logCollection.InsertOne(ctx, execLog)
		results = append(results, execLog)

		wf.Nodes[i].Status = execLog.Status
	}

	// Final workflow status
	wf.Status = "completed"
	collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": wf})

	c.JSON(http.StatusOK, gin.H{
		"workflow": wf.ID.Hex(),
		"status":   wf.Status,
		"results":  results,
	})
}

// ---------------- SAVE WORKFLOW STRUCTURE ----------------

func SaveWorkflowStructure(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var body struct {
		Nodes       []models.Node       `json:"nodes"`
		Connections []models.Connection `json:"connections"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"nodes":       body.Nodes,
		"connections": body.Connections,
	}}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save structure"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workflow structure updated",
		"id":      id,
	})
}
