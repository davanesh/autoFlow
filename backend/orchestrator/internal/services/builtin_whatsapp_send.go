package services

import (
	"errors"
	"fmt"

  wapp "github.com/Davanesh/auto-orchestrator/internal/executors"

)

func init() {
	RegisterExecutor("whatsapp_send", &WhatsAppSendExecutor{})
}

type WhatsAppSendExecutor struct{}

func (e *WhatsAppSendExecutor) Execute(n *ExecNode, g *ExecGraph) (string, error) {
	n.Status = "running"

	to := fmt.Sprintf("%v", n.Data["to"])
	if to == "" {
		n.Status = "failed"
		return "", errors.New("missing 'to' in whatsapp_send node")
	}

	body := ""
	if v, ok := n.Data["output"]; ok {
		body = fmt.Sprintf("%v", v)
	}
	if body == "" {
		if v, ok := n.Data["input"]; ok {
			body = fmt.Sprintf("%v", v)
		}
	}

	if body == "" {
		body = "(empty message)"
	}

	_, err := wapp.ExecuteWhatsAppSendNode(to, "static", "", "", body)
	if err != nil {
		n.Status = "failed"
		return "", err
	}

	n.Status = "done"
	return "", nil
}
