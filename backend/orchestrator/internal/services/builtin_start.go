package services

import "log"

type StartExecutor struct{}

func (s *StartExecutor) Execute(node *ExecNode, g *ExecGraph) (string, error) {
	log.Printf("🟢 Start node: %s", node.Label)
	node.Status = "done"
	return "", nil
}

func init() {
	RegisterExecutor("start", &StartExecutor{})
}
