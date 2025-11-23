package executors

import (
	"errors"
	"log"
	"time"

	"github.com/Davanesh/auto-orchestrator/internal/services"
)

type StartExecutor struct{}
func (e *StartExecutor) Execute(n *services.ExecNode, g *services.ExecGraph) (string, error) {
	log.Println("🟢 Start:", n.Label)
	n.Status = "done"
	if len(n.Next) > 0 {
		return n.Next[0], nil
	}
	return "", nil
}

type TaskExecutor struct{}
func (e *TaskExecutor) Execute(n *services.ExecNode, g *services.ExecGraph) (string, error) {
	log.Println("🟡 Task:", n.Label)
	n.Status = "running"
	time.Sleep(1 * time.Second)
	n.Status = "done"
	if len(n.Next) > 0 {
		return n.Next[0], nil
	}
	return "", nil
}

type DecisionExecutor struct{}
func (e *DecisionExecutor) Execute(n *services.ExecNode, g *services.ExecGraph) (string, error) {
	log.Println("🟣 Decision:", n.Label)
	cond, _ := n.Data["condition"].(string)
	if cond == "yes" {
		return n.Next[0], nil
	}
	if len(n.Next) < 2 {
		return "", errors.New("decision requires 2 branches")
	}
	return n.Next[1], nil
}

func init() {
	services.RegisterExecutor("start", &StartExecutor{})
	services.RegisterExecutor("task", &TaskExecutor{})
	services.RegisterExecutor("decision", &DecisionExecutor{})
	services.RegisterExecutor("ai", &AIExecutor{})
}
