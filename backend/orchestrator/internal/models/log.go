package models

import "time"

type ExecutionLog struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	WorkflowID  string    `bson:"workflowId" json:"workflowId"`
	TaskName    string    `bson:"taskName" json:"taskName"`
	Status      string    `bson:"status" json:"status"`
	Timestamp   time.Time `bson:"timestamp" json:"timestamp"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
}
