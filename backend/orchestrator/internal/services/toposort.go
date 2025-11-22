package services

import (
	"fmt"
)

// TopoSortResult holds both:
// - Linear order of nodeIDs
// - Parallel execution layers
type TopoSortResult struct {
	Order  []string     // simple linear list
	Layers [][]string   // grouped by in-degree == 0 batches
}

// TopologicalSort runs Kahn's Algorithm on the built graph.
// Requires: Graph already built via BuildGraph().
func TopologicalSort(g *Graph) (*TopoSortResult, error) {
	if g == nil || !g.HasNodes() {
		return nil, fmt.Errorf("cannot run topological sort: graph is empty")
	}

	// Copy indegree because we'll mutate it
	inDeg := make(map[string]int)
	for id, d := range g.InDegree {
		inDeg[id] = d
	}

	// Queue: nodes with in-degree 0
	queue := []string{}
	for id := range g.Nodes {
		if inDeg[id] == 0 {
			queue = append(queue, id)
		}
	}

	if len(queue) == 0 {
		return nil, fmt.Errorf("no start nodes found (cycle likely or no node with indegree 0)")
	}

	result := &TopoSortResult{
		Order:  []string{},
		Layers: [][]string{},
	}

	// Kahn's Algorithm – with "layers"
	for len(queue) > 0 {
		layer := append([]string(nil), queue...) // clone layer
		result.Layers = append(result.Layers, layer)

		next := []string{} // next layer

		for _, nodeID := range queue {
			result.Order = append(result.Order, nodeID)

			// Visit children
			for _, child := range g.Adj[nodeID] {
				inDeg[child]--
				if inDeg[child] == 0 {
					next = append(next, child)
				}
			}
		}

		// Prepare for next iteration
		queue = next
	}

	// Validate: If Order size != total nodes → cycle exists
	if len(result.Order) != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected: graph cannot be sorted (try checking missing connections)")
	}

	return result, nil
}
