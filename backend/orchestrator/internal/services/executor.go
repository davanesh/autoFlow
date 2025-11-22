package services

import (
	"errors"
	"log"
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

		executor, err := GetExecutor(n.Type)
		if err != nil {
				return errors.New("no executor for node type: " + n.Type)
		}

		nextOverride, err = executor.Execute(n, g)

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
