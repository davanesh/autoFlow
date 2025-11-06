package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Workflow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"createdAt"`
	Status      string             `json:"status"`
	Tasks       []Task             `json:"tasks"`
}

type Task struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Status string                 `json:"status"`
	Config map[string]interface{} `json:"config,omitempty"`
}
