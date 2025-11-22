package services

import (
	"fmt"
	"strings"

	"github.com/Davanesh/auto-orchestrator/internal/models"
)

// Graph is the adjacency representation used by the orchestrator.
type Graph struct {
	Adj      map[string][]string   // nodeID -> list of children nodeIDs
	InDegree map[string]int        // nodeID -> number of incoming edges
	Nodes    map[string]models.Node // nodeID -> node struct
}

// BuildGraph builds an adjacency list and in-degree map from nodes + connections.
// It requires that each node has an identifier (CanvasID). If a node is missing an id,
// it returns an error so the caller can handle it.
func BuildGraph(nodes []models.Node, conns []models.Connection) (*Graph, error) {
	g := &Graph{
		Adj:      make(map[string][]string),
		InDegree: make(map[string]int),
		Nodes:    make(map[string]models.Node),
	}

	// Normalize nodes and ensure IDs exist
	for _, n := range nodes {
		// prefer CanvasID (frontend id), fallback to Data["id"] or Label (but prefer explicit)
		id := strings.TrimSpace(n.CanvasID)
		if id == "" {
			// if user used `id` json key instead, support it by checking Data or Label as last resort
			if alt, ok := n.Data["id"].(string); ok && strings.TrimSpace(alt) != "" {
				id = strings.TrimSpace(alt)
			}
		}
		if id == "" {
			return nil, fmt.Errorf("node missing id/canvasId: label=%q type=%q", n.Label, n.Type)
		}

		// set canonical id in nodes map and initialize structures
		g.Nodes[id] = n
		if _, ok := g.Adj[id]; !ok {
			g.Adj[id] = []string{}
		}
		if _, ok := g.InDegree[id]; !ok {
			g.InDegree[id] = 0
		}
	}

	// Walk connections and populate adjacency + indegree
	for _, c := range conns {
		src := strings.TrimSpace(c.Source)
		tgt := strings.TrimSpace(c.Target)

		// ignore empty connections (some saved documents had empty strings)
		if src == "" || tgt == "" {
			continue
		}

		// validate nodes exist
		if _, ok := g.Nodes[src]; !ok {
			return nil, fmt.Errorf("connection source node not found: %s", src)
		}
		if _, ok := g.Nodes[tgt]; !ok {
			return nil, fmt.Errorf("connection target node not found: %s", tgt)
		}

		g.Adj[src] = append(g.Adj[src], tgt)
		// ensure target has indegree key
		if _, ok := g.InDegree[tgt]; !ok {
			g.InDegree[tgt] = 0
		}
		g.InDegree[tgt]++
	}

	return g, nil
}

// StartNodes returns list of node IDs with in-degree == 0 (good candidates to start execution).
// If no node has in-degree 0, the caller likely has a cycle or disconnected graph.
func (g *Graph) StartNodes() []string {
	res := []string{}
	for id := range g.Nodes {
		if g.InDegree[id] == 0 {
			res = append(res, id)
		}
	}
	return res
}

// HasNodes returns true if the graph contains any nodes.
func (g *Graph) HasNodes() bool {
	return len(g.Nodes) > 0
}

// DebugString returns a compact debug representation of the graph (adj + indegree).
func (g *Graph) DebugString() string {
	var sb strings.Builder
	sb.WriteString("Graph Debug:\n")
	for id := range g.Nodes {
		sb.WriteString(fmt.Sprintf("- %s (in=%d) -> %v\n", id, g.InDegree[id], g.Adj[id]))
	}
	return sb.String()
}
