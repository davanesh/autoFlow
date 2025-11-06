package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Davanesh/auto-orchestrator/internal/db"
	"github.com/Davanesh/auto-orchestrator/internal/models"
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
		log.Println("❌ Invalid ID format:", id)
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
		log.Println("❌ Update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		log.Println("⚠️ No workflow found for ID:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	log.Println("✅ Workflow status updated:", id, "->", body.Status)

	c.JSON(http.StatusOK, gin.H{
		"message": "Workflow status updated successfully",
		"id":      id,
		"status":  body.Status,
	})
}

// ---------------- GET ALL WORKFLOWS ----------------

func GetWorkflows(c *gin.Context) {
	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("📦 Fetching all workflows from MongoDB...")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("❌ Error fetching workflows:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var workflows []models.Workflow
	if err := cursor.All(ctx, &workflows); err != nil {
		log.Println("❌ Cursor decode error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Found %d workflows\n", len(workflows))
	c.JSON(http.StatusOK, workflows)
}

// ---------------- CREATE WORKFLOW ----------------

func CreateWorkflow(c *gin.Context) {
	var wf models.Workflow
	if err := c.BindJSON(&wf); err != nil {
		log.Println("❌ Invalid JSON for workflow:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wf.CreatedAt = time.Now()
	wf.Status = "draft"

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("🧠 Inserting new workflow into MongoDB:", wf.Name)

	result, err := collection.InsertOne(ctx, wf)
	if err != nil {
		log.Println("❌ Insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	wf.ID = result.InsertedID.(primitive.ObjectID)

	log.Println("✅ Workflow inserted successfully with ID:", wf.ID.Hex())

	c.JSON(http.StatusCreated, wf)
}

// ---------------- RUN WORKFLOW ----------------

func RunWorkflow(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("❌ Invalid workflow ID:", id)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("⚙️ Fetching workflow to execute:", id)

	var wf models.Workflow
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&wf)
	if err != nil {
		log.Println("❌ Workflow not found:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	results := []map[string]string{}
	logCollection := db.GetCollection("execution_logs")

	for i, task := range wf.Tasks {
		log.Printf("🚀 Running task %d: %s\n", i+1, task.Name)

		// mark as running
		wf.Tasks[i].Status = "running"
		_, _ = logCollection.InsertOne(ctx, models.ExecutionLog{
			WorkflowID:  wf.ID.Hex(),
			TaskName:    task.Name,
			Status:      "running",
			Timestamp:   time.Now(),
			Description: "Task started",
		})

		time.Sleep(1 * time.Second) // simulate processing

		// mark as completed
		wf.Tasks[i].Status = "completed"
		_, _ = logCollection.InsertOne(ctx, models.ExecutionLog{
			WorkflowID:  wf.ID.Hex(),
			TaskName:    task.Name,
			Status:      "completed",
			Timestamp:   time.Now(),
			Description: "Task completed successfully",
		})

		results = append(results, map[string]string{
			"task":   task.Name,
			"status": "completed",
		})

		log.Printf("✅ Task %d completed: %s\n", i+1, task.Name)
	}

	// update overall workflow
	wf.Status = "completed"
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": wf})
	if err != nil {
		log.Println("❌ Error updating workflow status:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("🎯 Workflow completed successfully:", wf.ID.Hex())

	c.JSON(http.StatusOK, gin.H{
		"workflowId": wf.ID.Hex(),
		"status":     wf.Status,
		"results":    results,
	})
}
