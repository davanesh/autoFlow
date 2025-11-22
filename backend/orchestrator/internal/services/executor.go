package services

import (
	"errors"
	"log"
	"time"
)


/*
----------------------------------------------------
    EXECUTION ENGINE
----------------------------------------------------
*/

func RunWorkflow(g *ExecGraph) error {
	if g == nil {
		return errors.New("nil graph")
	}
	if g.Start == "" {
		return errors.New("no start node defined")
	}

	log.Println("🚀 Starting workflow execution (new engine)")
	current := g.Start

	for {
		n, ok := g.Nodes[current]
		if !ok {
			return errors.New("node not found: " + current)
		}

		var nextOverride string
		var err error

		// Pick node type
		switch n.Type {
		case "start":
			err = executeStart(n)

		case "task":
			err = executeTask(n)

		case "decision":
			nextOverride, err = executeDecision(n)

		default:
			return errors.New("unknown node type: " + n.Type)
		}

		if err != nil {
			n.Status = "failed"
			log.Printf("❌ Node failed: %s (%v)", n.ID, err)
			return err
		}

		// If decision returned a target → follow it
		if nextOverride != "" {
			current = nextOverride
			continue
		}

		// No more next nodes → workflow ends
		if len(n.Next) == 0 {
			log.Println("🏁 Workflow complete!")
			return nil
		}

		// Non-decision nodes should only have 1 next
		if len(n.Next) > 1 {
			return errors.New("node has multiple next branches but is not a decision: " + n.ID)
		}

		current = n.Next[0]
	}
}

/*
----------------------------------------------------
    NODE EXECUTORS
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
	time.Sleep(1 * time.Second) // simulate work

	n.Status = "done"
	log.Printf("✅ Task completed: %s", n.Label)
	return nil
}

func executeDecision(n *ExecNode) (string, error) {
	log.Printf("🟣 Decision: %s", n.Label)

	cond, ok := n.Data["condition"].(string)
	if !ok {
		return "", errors.New("decision node missing 'condition'")
	}

	if len(n.Next) < 2 {
		return "", errors.New("decision node must have 2+ branches")
	}

	if cond == "yes" {
		log.Println("➡️ Decision: YES branch")
		return n.Next[0], nil
	}

	log.Println("➡️ Decision: NO branch")
	return n.Next[1], nil
}
