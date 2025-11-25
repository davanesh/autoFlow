package services

import (
	"strconv"
	"time"
)

func init() {
	RegisterExecutor("wait", &WaitExecutor{})
}

type WaitExecutor struct{}

func (e *WaitExecutor) Execute(n *ExecNode, g *ExecGraph) (string, error) {
	n.Status = "running"

	secs := 0
	if v, ok := n.Data["waitSeconds"]; ok {
		switch t := v.(type) {
		case float64:
			secs = int(t)
		case int:
			secs = t
		case string:
			secs, _ = strconv.Atoi(t)
		}
	}

	time.Sleep(time.Duration(secs) * time.Second)
	n.Status = "done"
	return "", nil
}
