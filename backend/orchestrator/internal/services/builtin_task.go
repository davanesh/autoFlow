package services

import (
	"log"
	"time"
)

type TaskExecutor struct{}

func (t *TaskExecutor) Execute(node *ExecNode, g *ExecGraph) (string, error) {
	log.Printf("🟡 Task node: %s", node.Label)
	node.Status = "running"

	// Simulated work
	if v, ok := node.Data["sleepMs"]; ok {
		if ms, ok2 := v.(float64); ok2 {
			time.Sleep(time.Duration(int(ms)) * time.Millisecond)
		}
	} else {
		time.Sleep(500 * time.Millisecond)
	}

	node.Status = "done"
	log.Printf("✅ Task completed: %s", node.Label)
	return "", nil
}

func init() {
	RegisterExecutor("task", &TaskExecutor{})
}
