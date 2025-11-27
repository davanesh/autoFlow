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

func RegisterWorkflowRoutes(r *gin.Engine) {
	r.GET("/workflows", GetWorkflows)
	r.GET("/workflows/:id", GetWorkflowByID)
	r.POST("/workflows", CreateWorkflow)
	r.PUT("/workflows/:id", UpdateWorkflowStatus)
	r.POST("/workflows/:id/run", RunWorkflow)
	r.PUT("/workflows/:id/structure", SaveWorkflowStructure)
}

// -----------------------------------------------------
// UPDATE STATUS
// -----------------------------------------------------

func UpdateWorkflowStatus(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Status string `json:"status"`
	}

	if err := c.BindJSON(&body); err != nil {
		log.Println("❌ JSON error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, _ := primitive.ObjectIDFromHex(id)

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(ctx, bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"status": body.Status}})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated",
		"status":  body.Status,
	})
}

// -----------------------------------------------------
// GET ALL WORKFLOWS
// -----------------------------------------------------

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
	cursor.All(ctx, &workflows)

	c.JSON(http.StatusOK, workflows)
}

// -----------------------------------------------------
// GET WORKFLOW BY ID
// -----------------------------------------------------

func GetWorkflowByID(c *gin.Context) {
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wf models.Workflow
	err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&wf)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, wf)
}

// -----------------------------------------------------
// CREATE WORKFLOW
// -----------------------------------------------------

func CreateWorkflow(c *gin.Context) {
	var wf models.Workflow

	if err := c.BindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	wf.CreatedAt = time.Now()
	wf.Status = "draft"

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, _ := collection.InsertOne(ctx, wf)
	wf.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, wf)
}

// -----------------------------------------------------
// RUN WORKFLOW ENGINE
// -----------------------------------------------------

func RunWorkflow(c *gin.Context) {
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)

	collection := db.GetCollection("workflows")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var wf models.Workflow
	err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&wf)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	graph := services.ExecGraph{
		Nodes: map[string]*services.ExecNode{},
		Start: "",
		RunID: primitive.NewObjectID().Hex(),
	}

	// ---------------------------------------------------------
	// ① FIX NODE IDS (canvasId)
	// ---------------------------------------------------------

	for i := range wf.Nodes {
		node := &wf.Nodes[i]

		canvas := node.CanvasID
		if canvas == "" && node.LegacyID != "" {
			canvas = node.LegacyID
		}
		if canvas == "" {
			canvas = strings.ToLower(strings.ReplaceAll(node.Label, " ", ""))
			log.Println("⚠️ Auto-generated ID for node:", node.Type, "->", canvas)
		}

		node.CanvasID = canvas

		graph.Nodes[canvas] = &services.ExecNode{
			ID:     canvas,
			Type:   normalizeNodeType(node.Type),
			Label:  node.Label,
			Data:   node.Data,
			Status: "pending",
			Next:   []string{},
		}

		if strings.ToLower(node.Type) == "start" {
			graph.Start = canvas
		}
	}

	// ---------------------------------------------------------
	// ② AUTO-MAP BAD CONNECTION IDs → VALID IDs
	// ---------------------------------------------------------

	idMap := map[string]string{}
	for _, node := range wf.Nodes {
		switch node.Type {
		case "whatsapp_wait":
			idMap["wait1"] = node.CanvasID
		case "whatsapp_static_reply":
			idMap["static1"] = node.CanvasID
		case "whatsapp_send":
			idMap["send1"] = node.CanvasID
		}
		idMap[node.Label] = node.CanvasID
	}

	// ---------------------------------------------------------
	// ③ APPLY CONNECTIONS
	// ---------------------------------------------------------

	for _, conn := range wf.Connections {

		src := conn.Source
		tgt := conn.Target

		if v, ok := idMap[src]; ok {
			src = v
		}
		if v, ok := idMap[tgt]; ok {
			tgt = v
		}

		if node, exists := graph.Nodes[src]; exists {
			node.Next = append(node.Next, tgt)
		} else {
			log.Printf("⚠️ Invalid connection source: %s -> %s", src, tgt)
		}
	}

	if graph.Start == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No start node defined"})
		return
	}

	// ---------------------------------------------------------
	// ④ RUN ENGINE
	// ---------------------------------------------------------

	log.Println("🔥 Running NEW EXECUTION ENGINE...")

	err = services.RunWorkflow(&graph)
	if err != nil {
		log.Println("❌ Engine error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---------------------------------------------------------
	// ⑤ SAVE NODE RESULTS BACK TO DB
	// ---------------------------------------------------------

	for i := range wf.Nodes {
		n := &wf.Nodes[i]
		if updated, ok := graph.Nodes[n.CanvasID]; ok {
			n.Status = updated.Status
			n.Data = updated.Data
		}
	}

	wf.Status = "completed"

	collection.UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"nodes": wf.Nodes, "status": wf.Status}},
	)

	c.JSON(http.StatusOK, gin.H{
		"workflowId": wf.ID.Hex(),
		"status":     wf.Status,
		"nodes":      wf.Nodes,
	})
}

// -----------------------------------------------------
// NORMALIZE NODE TYPE
// -----------------------------------------------------

func normalizeNodeType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))

	switch t {
	case "start":
		return "start"
	case "task":
		return "task"
	case "decision":
		return "decision"
	case "ai":
		return "ai"
	case "wait":
		return "wait"
	case "whatsapp_wait":
		return "whatsapp_wait"
	case "whatsapp_static_reply":
		return "whatsapp_static_reply"
	case "whatsapp_send":
		return "whatsapp_send"
	}

	return "task"
}

// -----------------------------------------------------
// SAVE WORKFLOW STRUCTURE
// -----------------------------------------------------

func SaveWorkflowStructure(c *gin.Context) {
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)

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

	collection.UpdateOne(ctx, bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"nodes":       body.Nodes,
			"connections": body.Connections,
		}},
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Structure saved",
		"id":      id,
	})
}
