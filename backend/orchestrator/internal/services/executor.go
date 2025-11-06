package services

import (
	"log"
	"time"
)

type Task struct {
	Name   string
	Status string
}

func ExecuteTask(t *Task) {
	log.Printf("🟡 Starting task: %s", t.Name)
	t.Status = "running"
	time.Sleep(2 * time.Second) // Simulate work
	t.Status = "done"
	log.Printf("✅ Completed task: %s", t.Name)
}
