package services

import (
	"strconv"
	wapp "github.com/Davanesh/auto-orchestrator/internal/executors"

)

func init() {
	RegisterExecutor("whatsapp_wait", &WhatsAppWaitExecutor{})
}

type WhatsAppWaitExecutor struct{}

func (e *WhatsAppWaitExecutor) Execute(n *ExecNode, g *ExecGraph) (string, error) {
	n.Status = "running"

	timeout := 0
	if v, ok := n.Data["timeoutSeconds"]; ok {
		switch t := v.(type) {
		case float64:
			timeout = int(t)
		case int:
			timeout = t
		case string:
			timeout, _ = strconv.Atoi(t)
		}
	}

	msg, err := wapp.WaitForWhatsAppMessage(g.RunID, n.ID, timeout)
	if err != nil {
		n.Status = "failed"
		return "", err
	}

	if n.Data == nil {
		n.Data = map[string]interface{}{}
	}

	n.Data["input"] = msg
	n.Status = "done"
	return "", nil
}
