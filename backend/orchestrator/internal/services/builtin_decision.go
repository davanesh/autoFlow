package services

import (
	"errors"
	"log"
)

type DecisionExecutor struct{}

func (d *DecisionExecutor) Execute(node *ExecNode, g *ExecGraph) (string, error) {
	log.Printf("🟣 Decision node: %s", node.Label)
	node.Status = "running"

	cond, ok := node.Data["condition"]
	if !ok {
		return "", errors.New("decision node missing condition")
	}

	// Normalize to string
	condStr := ""
	switch v := cond.(type) {
	case string:
		condStr = v
	case bool:
		if v {
			condStr = "true"
		} else {
			condStr = "false"
		}
	default:
		return "", errors.New("invalid condition type")
	}

	node.Status = "done"

	// Yes / True branch = Next[0]
	if condStr == "true" || condStr == "yes" || condStr == "1" {
		if len(node.Next) > 0 {
			return node.Next[0], nil
		}
		return "", errors.New("missing true branch")
	}

	// No / False branch = Next[1]
	if len(node.Next) > 1 {
		return node.Next[1], nil
	}

	return "", errors.New("missing false branch")
}

func init() {
	RegisterExecutor("decision", &DecisionExecutor{})
}
