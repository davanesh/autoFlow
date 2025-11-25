package executors

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// NOTE: adapt imports and function names to your project's patterns.
// This file provides:
// - webhook handler: HandleWhatsAppWebhook
// - Waiter registration: WaitForWhatsAppMessage
// - Send helper: SendWhatsAppMessage
// - Execute wait/send logic that your orchestrator can call.

// In-memory waiter registry: key = runID + ":" + nodeID
var waiters = struct {
	m sync.Map // map[string]chan string
}{}

func waiterKey(runID, nodeID string) string {
	return fmt.Sprintf("%s:%s", runID, nodeID)
}

// RegisterWaiter registers a channel and waits for incoming message or timeout.
// If timeoutSeconds == 0, it waits indefinitely (be careful).
func RegisterWaiter(runID, nodeID string, timeoutSeconds int) (chan string, error) {
	key := waiterKey(runID, nodeID)
	ch := make(chan string, 1)
	// store channel
	waiters.m.Store(key, ch)
	// return channel for caller to wait on
	if timeoutSeconds > 0 {
		go func() {
			<-time.After(time.Duration(timeoutSeconds) * time.Second)
			// send empty string or close to avoid blocking forever
			if val, ok := waiters.m.Load(key); ok {
				if c, ok2 := val.(chan string); ok2 {
					select {
					case c <- "__TIMEOUT__":
					default:
					}
					// cleanup
					waiters.m.Delete(key)
				}
			}
		}()
	}
	return ch, nil
}

// deliverMessageToWaiter delivers incoming text to registered waiter if exists.
func deliverMessageToWaiter(runID, nodeID, text string) bool {
	key := waiterKey(runID, nodeID)
	if val, ok := waiters.m.Load(key); ok {
		if ch, ok2 := val.(chan string); ok2 {
			select {
			case ch <- text:
			default:
			}
			waiters.m.Delete(key)
			return true
		}
	}
	return false
}

// Webhook payload handling (Twilio sends form values)
type twilioWebhookPayload struct {
	From string
	To   string
	Body string
	// Twilio also sends other fields (MessageSid etc). Add if needed.
}

// Allowed sender check (only your number for now)
func isAllowedSender(from string) bool {
	allowed := os.Getenv("ALLOWED_WHATSAPP_NUMBER") // e.g. whatsapp:+91999...
	return allowed == "" || strings.Contains(from, allowed) || from == allowed
}

// HandleWhatsAppWebhook is the HTTP handler to receive Twilio's webhook POSTs.
// Mount as "/webhook/whatsapp".
func HandleWhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("parse form err:", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	payload := twilioWebhookPayload{
		From: r.FormValue("From"),
		To:   r.FormValue("To"),
		Body: r.FormValue("Body"),
	}
	log.Printf("WhatsApp incoming from=%s body=%s\n", payload.From, payload.Body)

	if !isAllowedSender(payload.From) {
		log.Printf("sender not allowed: %s\n", payload.From)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// We expect the message to include a runID and nodeID to route to a waiter.
	// Two options:
	// 1) You design to send "run:RUNID node:NODEID msg:..." from your phone, OR
	// 2) If there's only one active waiter per run, we can route by runID only.
	// For now, we support both:
	// - If body contains "run:<runID> node:<nodeID> message:<payload>" we parse it.
	// - Otherwise, we attempt deliver to any waiter matching runID found in body.
	// Simple parsing:
	text := payload.Body

	// Try to parse "run:RUN node:NODE msg:...". Case-insensitive
	textLower := strings.ToLower(text)
	runID := ""
	nodeID := ""

	runRe := regexp.MustCompile(`run[:=]\s*([^\s]+)`)
	nodeRe := regexp.MustCompile(`node[:=]\s*([^\s]+)`)
	if m := runRe.FindStringSubmatch(textLower); len(m) > 1 {
		runID = m[1]
	}
	if m := nodeRe.FindStringSubmatch(textLower); len(m) > 1 {
		nodeID = m[1]
	}

	// If both available, deliver directly.
	if runID != "" && nodeID != "" {
		delivered := deliverMessageToWaiter(runID, nodeID, payload.Body)
		if delivered {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "Delivered")
			return
		}
		// not delivered, return 404 so Twilio can retry (or 200 if you prefer)
		http.Error(w, "no waiter", http.StatusNotFound)
		return
	}

	// If no run/node provided, try to find any waiter where runID or nodeID is embedded.
	// Iterate over waiters keys and try fuzzy match by run or by containing phone as key.
	delivered := false
	waiters.m.Range(func(k, v interface{}) bool {
		keyStr := k.(string) // runID:nodeID
		parts := strings.SplitN(keyStr, ":", 2)
		if len(parts) != 2 {
			return true // continue
		}
		// If body contains runID or nodeID, deliver.
		if parts[0] != "" && strings.Contains(strings.ToLower(textLower), strings.ToLower(parts[0])) {
			if ch, ok := v.(chan string); ok {
				select {
				case ch <- payload.Body:
				default:
				}
				waiters.m.Delete(keyStr)
				delivered = true
				return false // stop iteration
			}
		}
		if parts[1] != "" && strings.Contains(strings.ToLower(textLower), strings.ToLower(parts[1])) {
			if ch, ok := v.(chan string); ok {
				select {
				case ch <- payload.Body:
				default:
				}
				waiters.m.Delete(keyStr)
				delivered = true
				return false // stop iteration
			}
		}
		return true
	})
	if delivered {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Delivered fuzzy")
		return
	}

	// not delivered: just 200 OK to stop Twilio retries or 404 to force retry. We'll return 200.
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "No registered waiter")
}

// WaitNode execution: called by orchestrator when a Wait node runs.
// It registers a waiter and blocks until a message arrives.
// Return: incoming text or error.
func WaitForWhatsAppMessage(runID, nodeID string, timeoutSeconds int) (string, error) {
	ch, err := RegisterWaiter(runID, nodeID, timeoutSeconds)
	if err != nil {
		return "", err
	}
	// Wait for message
	msg := <-ch
	if msg == "__TIMEOUT__" {
		return "", errors.New("waiter timeout")
	}
	return msg, nil
}

// Static template with regex captures.
// regexPattern: optional, if empty just send template.
// template: supports ${1}, ${2} for capture groups.
// message: incoming message to match against regex
func BuildStaticReply(regexPattern, template, message string) (string, error) {
	if regexPattern == "" {
		// no regex, return template as-is (can also support placeholders like ${body})
		res := strings.ReplaceAll(template, "${body}", message)
		return res, nil
	}
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return "", err
	}
	matches := re.FindStringSubmatch(message)
	if matches == nil {
		// no match -> return template as-is or signal no match. We'll return template.
		res := strings.ReplaceAll(template, "${body}", message)
		return res, nil
	}
	// Replace ${n} placeholders
	out := template
	for i := 1; i < len(matches); i++ {
		placeholder := fmt.Sprintf("${%d}", i)
		out = strings.ReplaceAll(out, placeholder, matches[i])
	}
	// Support ${body}
	out = strings.ReplaceAll(out, "${body}", message)
	return out, nil
}

// SendWhatsAppMessage posts to Twilio to send WhatsApp message.
func SendWhatsAppMessage(to, body string) error {
	accountSid := os.Getenv("TWILIO_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_WHATSAPP_FROM") // e.g. whatsapp:+1415...
	if accountSid == "" || authToken == "" || from == "" {
		return errors.New("twilio env vars not set")
	}
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)
	data := url.Values{}
	// Twilio WhatsApp requires "whatsapp:+<number>"
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}
	data.Set("To", to)
	data.Set("From", from)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Twilio send ok. resp=%s\n", string(b))
		return nil
	}
	return fmt.Errorf("twilio error status=%d body=%s", resp.StatusCode, string(b))
}

// Example orchestrator-facing helper: ExecuteWhatsAppSendNode
// `input` is the incoming message (if any) or previous node output.
// mode: "static" or "ai". For "static" regexPattern + template used. For "ai", call internal ai.
func ExecuteWhatsAppSendNode(to string, mode string, regexPattern, template, input string) (string, error) {
	out := ""
	if mode == "static" {
		r, err := BuildStaticReply(regexPattern, template, input)
		if err != nil {
			return "", err
		}
		out = r
	} else if mode == "ai" {
		// call internal AI node
		// IMPORTANT: replace this with actual call to your AI node package.
		// We assume a package "ainode" with Process(input string) (string, error)
		reply, err := internalAiProcess(input)
		if err != nil {
			return "", err
		}
		out = reply
	} else {
		return "", errors.New("unknown mode")
	}

	// send out
	if err := SendWhatsAppMessage(to, out); err != nil {
		return "", err
	}
	return out, nil
}

// internalAiProcess is a placeholder that calls your internal AI Node.
// Replace with your actual code reference.
func internalAiProcess(in string) (string, error) {
	// Example: call package aINode.UsingProcess
	// return ainode.Process(in)
	// For now returning a stub.
	// TODO: swap with your real AI Node integration.
	if procFunc := getAIFunc(); procFunc != nil {
		return procFunc(in)
	}
	// fallback stub:
	return fmt.Sprintf("AI reply (stub): I heard \"%s\"", in), nil
}

// getAIFunc tries to get a function from a package that may be present in your project.
// You can replace this entirely with direct call to your AI node.
func getAIFunc() func(string) (string, error) {
	// replace: return ainode.Process
	return nil
}

// Gin wrapper for webhook handler
func HandleWhatsAppWebhookGin(c *gin.Context) {
	HandleWhatsAppWebhook(c.Writer, c.Request)
}
