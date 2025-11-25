package services

// node.go
// Core node types & small helpers

type NodeType string

const (
	NodeTypeStart    NodeType = "start"
	NodeTypeTask     NodeType = "task"
	NodeTypeDecision NodeType = "decision"
)

type ExecNode struct {
	ID     string
	Type   string
	Label  string
	Data   map[string]interface{}
	Status string
	Next   []string // adjacency
}

// ExecGraph already used by your api
type ExecGraph struct {
	Nodes map[string]*ExecNode
	Start string
	RunID string
}
