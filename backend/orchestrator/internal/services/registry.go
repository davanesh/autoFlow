package services

import "fmt"

// NodeExecutor executes a node and returns next node id (or empty if engine should use Next[]).
// For decision nodes it returns the next node id to jump to.
// For normal nodes it returns "" meaning: engine will pick node.Next[0] (must exist).
type NodeExecutor interface {
	Execute(node *ExecNode, g *ExecGraph) (next string, err error)
}

// simple registry
var executors = map[string]NodeExecutor{}

func RegisterExecutor(nodeType string, exec NodeExecutor) {
	executors[nodeType] = exec
}

func GetExecutor(nodeType string) (NodeExecutor, error) {
	if e, ok := executors[nodeType]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("no executor registered for node type: %s", nodeType)
}
