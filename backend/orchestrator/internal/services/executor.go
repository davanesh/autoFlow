package services

import (
	"errors"
	"fmt"
	"log"
	"time"
)

/*
----------------------------------------------------
    NODE MODELS (MATCHING BACKEND WORKFLOW FORMAT)
----------------------------------------------------
*/

type ExecNode struct {
	ID       string
	Type     string
	Label    string
	Data     map[string]interface{}
	Status   string
	Next     []string // adjacency list
}

type ExecGraph struct {
	Nodes map[string]*ExecNode
	Start string
}

/*
----------------------------------------------------
    MAIN EXECUTION ENGINE
----------------------------------------------------
*/

func RunWorkflow(g *ExecGraph) error {
	if g.Start == "" {
		return errors.New("❌ no start node defined in graph")
	}

	log.Println("🚀 Starting workflow execution")
	log.Printf("▶️  Entry node: %s", g.Start)

	current := g.Start

	for {
		node, ok := g.Nodes[current]
		if !ok {
			return fmt.Errorf("❌ node '%s' not found in graph", current)
		}

		var err error

		switch node.Type {
		case "start":
			err = executeStart(node)

		case "task":
			err = executeTask(node)

		case "decision":
			current, err = executeDecision(node)
			if err != nil {
				return err
			}
			continue // skip normal next-node handling

		default:
			return fmt.Errorf("❌ unknown node type: %s", node.Type)
		}

		if err != nil {
			return err
		}

		// Normal sequential move
		if len(node.Next) == 0 {
			log.Println("🏁 Workflow complete!")
			return nil
		}

		if len(node.Next) > 1 {
			return fmt.Errorf("❌ multiple next paths for non-decision node: %s", node.ID)
		}

		current = node.Next[0]
	}
}

/*
----------------------------------------------------
    EXECUTE NODE TYPES
----------------------------------------------------
*/

func executeStart(n *ExecNode) error {
	log.Printf("🟢 Start: %s", n.Label)
	n.Status = "done"
	return nil
}

func executeTask(n *ExecNode) error {
	log.Printf("🟡 Running task: %s", n.Label)

	n.Status = "running"

	// Simulate work (later replaced by Lambda / Python sandbox / Actions)
	time.Sleep(1 * time.Second)

	n.Status = "done"
	log.Printf("✅ Task completed: %s", n.Label)
	return nil
}

func executeDecision(n *ExecNode) (string, error) {
	log.Printf("🟣 Decision: %s", n.Label)

	// Example decision input (you replace with your actual logic)
	cond, ok := n.Data["condition"].(string)
	if !ok {
		return "", errors.New("❌ decision node missing condition field")
	}

	if len(n.Next) < 2 {
		return "", errors.New("❌ decision node must have at least 2 branches")
	}

	// Simple demo logic
	if cond == "yes" {
		log.Println("➡️ Decision: YES branch")
		return n.Next[0], nil
	}

	log.Println("➡️ Decision: NO branch")
	return n.Next[1], nil
}
