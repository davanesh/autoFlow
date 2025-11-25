package services

import (
	"fmt"
	wapp "github.com/Davanesh/auto-orchestrator/internal/executors"
)

func init() {
	RegisterExecutor("whatsapp_static_reply", &WhatsAppStaticReplyExecutor{})
}

type WhatsAppStaticReplyExecutor struct{}

func (e *WhatsAppStaticReplyExecutor) Execute(n *ExecNode, g *ExecGraph) (string, error) {
	n.Status = "running"

	input := fmt.Sprintf("%v", n.Data["input"])
	regex := fmt.Sprintf("%v", n.Data["match_regex"])
	template := fmt.Sprintf("%v", n.Data["reply_template"])

	out, err := wapp.BuildStaticReply(regex, template, input)
	if err != nil {
		n.Status = "failed"
		return "", err
	}

	n.Data["output"] = out
	n.Status = "done"
	return "", nil
}
