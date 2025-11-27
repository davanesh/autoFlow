package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Workflow represents a saved workflow. It contains both the older Tasks slice
// (kept for backward compatibility) and the canvas-friendly Nodes & Connections.
type Workflow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	Status      string             `bson:"status" json:"status"`

	// Backwards-compatible task list (your earlier code used this)
	Tasks []Task `bson:"tasks,omitempty" json:"tasks,omitempty"`

	// Canvas-friendly representation (nodes & edges)
	Nodes       []Node       `bson:"nodes,omitempty" json:"nodes,omitempty"`
	Connections []Connection `bson:"connections,omitempty" json:"connections,omitempty"`
}

// Task is the older unit of work representation (kept for compatibility)
type Task struct {
	Name   string                 `bson:"name" json:"name"`
	Type   string                 `bson:"type" json:"type"`
	Status string                 `bson:"status" json:"status"`
	Config map[string]interface{} `bson:"config,omitempty" json:"config,omitempty"`
}

// Node represents a canvas node (frontend ID is preserved in CanvasID)
type Node struct {
    // CanvasID is the frontend id we normally use. It maps to BSON field "canvasId".
    CanvasID   string                 `bson:"canvasId,omitempty" json:"id,omitempty"`

    // LegacyID will capture nodes stored with "id" in BSON (older save variations).
    // It's ignored for JSON output but read from BSON so we can migrate at runtime.
    LegacyID   string                 `bson:"id,omitempty" json:"-"`

    Type       string                 `bson:"type" json:"type"`
    Label      string                 `bson:"label,omitempty" json:"label,omitempty"`
    Position   map[string]float64     `bson:"position,omitempty" json:"position,omitempty"`
    Data       map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
    Status     string                 `bson:"status,omitempty" json:"status,omitempty"`
    LambdaName string                 `bson:"lambdaName,omitempty" json:"lambdaName,omitempty"`
}

// Connection (edge) between nodes
type Connection struct {
	// CanvasID is the frontend connection id (if available)
	CanvasID string `bson:"canvasId,omitempty" json:"id,omitempty"`

	Source string                 `bson:"source" json:"source"`
	Target string                 `bson:"target" json:"target"`
	Meta   map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}
