package api

import (
	"context"
	"log"
	"net/http"
	"strings"
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

// ---------------- RUN WORKFLOW (NEW ENGINE) ----------------

func RunWorkflow(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Load workflow from DB
	var wf models.Workflow
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&wf); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// -------------------------------
	// 1) Convert Workflow → ExecGraph
	// -------------------------------

	graph := services.ExecGraph{
		Nodes: map[string]*services.ExecNode{},
		Start: "",
	}

	// Build nodes
graph.Start = "" // reset
for _, n := range wf.Nodes {
  execNode := &services.ExecNode{
    ID:     n.CanvasID,
    Type:   normalizeNodeType(n.Type),
    Label:  n.Label,
    Data:   n.Data,
    Status: "pending",
		Next:   []string{},
  }
  graph.Nodes[n.CanvasID] = execNode
    // Flexible check
  if n.Type != "" && (strings.ToLower(n.Type) == "start") {        
		graph.Start = n.CanvasID
  }
}


	// Build edges (connections)
	for _, c2 := range wf.Connections {
		if node, exists := graph.Nodes[c2.Source]; exists {
			node.Next = append(node.Next, c2.Target)
		}
	}

	if graph.Start == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No start node defined"})
		return
	}

	// -------------------------------
	// 2) Run workflow engine
	// -------------------------------

	log.Println("🔥 Running NEW EXECUTION ENGINE...")

	err = services.RunWorkflow(&graph)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// -------------------------------
	// 3) Sync results back to DB
	// -------------------------------

	for i := range wf.Nodes {
		if updated, ok := graph.Nodes[wf.Nodes[i].CanvasID]; ok {
			wf.Nodes[i].Status = updated.Status
		}
	}

	wf.Status = "completed"

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{
		"$set": bson.M{
			"nodes":  wf.Nodes,
			"status": wf.Status,
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save updates"})
		return
	}

	// -------------------------------
	// 4) Response
	// -------------------------------

	c.JSON(http.StatusOK, gin.H{
		"workflowId": wf.ID.Hex(),
		"status":     wf.Status,
		"nodes":      wf.Nodes,
	})
}

// Normalizes frontend type → backend type
func normalizeNodeType(t string) string {
	switch t {
	case "Start", "start":
		return "start"
	case "Task", "task", "LambdaTask":
		return "task"
	case "Decision", "decision":
		return "decision"
	}
	return "task"
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
