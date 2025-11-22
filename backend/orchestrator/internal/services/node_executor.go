package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Davanesh/auto-orchestrator/internal/models"
)

// NodeExecutor implements NodeRunner (from executor.go) and runs nodes.
// It supports types:
//  - "lambda" (simulated invocation; use TODO to plug real AWS invoker)
//  - "http" (perform HTTP request; fields: url, method, headers, body)
//  - "ai" (POST to ai-engine local endpoint; fields: model, prompt, params)
//  - "decision" (evaluate a simple condition present in node.Data["condition"])
//
// Retry/backoff behavior:
//  - read node.Data["retries"] (int) default 1
//  - read node.Data["timeoutMs"] (int) per-node timeout in ms (default 30000)
//  - exponential backoff base 500ms
type NodeExecutor struct {
	// HTTP client reused
	httpClient *http.Client
	// Optionally you can provide a real Lambda invoker here.
	// LambdaInvoker LambdaInvokerInterface // TODO: define & set if you have an AWS lambda wrapper
}

func NewNodeExecutor() *NodeExecutor {
	return &NodeExecutor{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// Run executes node. Conforms to NodeRunner interface.
func (ne *NodeExecutor) Run(ctx context.Context, node models.Node, input map[string]interface{}) (map[string]interface{}, error) {
	typ := strings.ToLower(node.Type)
	// Support both 'Lambda process-data' style and 'lambda' simple types: normalize
	if strings.HasPrefix(strings.ToLower(node.Type), "lambda") {
		typ = "lambda"
	}

	// read retries & timeout from node.Data
	retries := 1
	if r, ok := node.Data["retries"]; ok {
		switch v := r.(type) {
		case float64:
			retries = int(v)
		case int:
			retries = v
		}
	}
	if retries < 1 {
		retries = 1
	}

	timeoutMs := 30000
	if t, ok := node.Data["timeoutMs"]; ok {
		switch v := t.(type) {
		case float64:
			timeoutMs = int(v)
		case int:
			timeoutMs = v
		}
	}

	var lastErr error
	backoffBase := 500 * time.Millisecond

	for attempt := 1; attempt <= retries; attempt++ {
		// per-attempt context with timeout
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()

		var out map[string]interface{}
		var err error

		switch typ {
		case "lambda":
			out, err = ne.execLambda(attemptCtx, node, input)

		case "http":
			out, err = ne.execHTTP(attemptCtx, node, input)

		case "ai":
			out, err = ne.execAI(attemptCtx, node, input)

		case "decision":
			out, err = ne.execDecision(attemptCtx, node, input)

		default:
			// fallback: treat as simulated task
			out, err = ne.execSimulate(attemptCtx, node, input)
		}

		if err == nil {
			return out, nil
		}

		// failure -> maybe retry
		lastErr = err
		log.Printf("⚠️ Node %s attempt %d/%d failed: %v", node.CanvasID, attempt, retries, err)
		// backoff before next attempt (unless last attempt)
		if attempt < retries {
			sleep := time.Duration(attempt) * backoffBase
			select {
			case <-time.After(sleep):
				// continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("node %s failed after %d attempts: %w", node.CanvasID, retries, lastErr)
}

// ---- executor helpers ----

func (ne *NodeExecutor) execSimulate(ctx context.Context, node models.Node, input map[string]interface{}) (map[string]interface{}, error) {
	// tiny simulation: wait 200-600ms
	select {
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	out := map[string]interface{}{
		"note":     "simulated execution",
		"nodeId":   node.CanvasID,
		"nodeType": node.Type,
		"input":    input,
	}
	return out, nil
}

func (ne *NodeExecutor) execLambda(ctx context.Context, node models.Node, input map[string]interface{}) (map[string]interface{}, error) {
	// If you have a real Lambda invoker, plug it here.
	// For now we simulate (safe & fast) but include details from node.Data for realism.
	lambdaName := ""
	if v, ok := node.Data["lambdaName"].(string); ok {
		lambdaName = v
	}
	if lambdaName == "" {
		// try node.Data["functionName"]
		if v, ok := node.Data["functionName"].(string); ok {
			lambdaName = v
		}
	}
	// simulated payload response
	out := map[string]interface{}{
		"invokedFunction": lambdaName,
		"message":         fmt.Sprintf("simulated lambda '%s' executed", lambdaName),
		"input":           input,
	}
	// small sleep to simulate execution
	select {
	case <-time.After(400 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	// TODO: if you have real Lambda wrapper:
	// respPayload, err := ne.LambdaInvoker.Invoke(lambdaName, input)
	// parse respPayload -> return map/err
	return out, nil
}

func (ne *NodeExecutor) execHTTP(ctx context.Context, node models.Node, input map[string]interface{}) (map[string]interface{}, error) {
	urlStr := ""
	if v, ok := node.Data["url"].(string); ok {
		urlStr = v
	}
	if urlStr == "" {
		return nil, fmt.Errorf("http node missing 'url' in data")
	}
	method := "POST"
	if v, ok := node.Data["method"].(string); ok && v != "" {
		method = strings.ToUpper(v)
	}

	// body: prefer explicit body in node, else use input
	var bodyBytes []byte
	if b, ok := node.Data["body"]; ok && b != nil {
		// if it's string or map, marshal accordingly
		switch x := b.(type) {
		case string:
			bodyBytes = []byte(x)
		default:
			j, _ := json.Marshal(x)
			bodyBytes = j
		}
	} else {
		j, _ := json.Marshal(input)
		bodyBytes = j
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// headers
	if hdrs, ok := node.Data["headers"].(map[string]interface{}); ok {
		for k, v := range hdrs {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}
	// ensure JSON
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := ne.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respB, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// try to parse JSON
	var parsed interface{}
	if len(respB) > 0 {
		if err := json.Unmarshal(respB, &parsed); err != nil {
			parsed = string(respB)
		}
	}

	out := map[string]interface{}{
		"statusCode": resp.StatusCode,
		"body":       parsed,
	}
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return out, nil
}

func (ne *NodeExecutor) execAI(ctx context.Context, node models.Node, input map[string]interface{}) (map[string]interface{}, error) {
	// Basic behaviour: POST to AI engine (ai-engine) at localhost:9000 /predict (change as needed)
	aiURL := "http://localhost:9000/predict"
	if v, ok := node.Data["aiUrl"].(string); ok && v != "" {
		aiURL = v
	}

	payload := map[string]interface{}{
		"node":  node,
		"input": input,
	}
	if p, ok := node.Data["prompt"]; ok {
		payload["prompt"] = p
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", aiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ne.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respB, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed interface{}
	if len(respB) > 0 {
		if err := json.Unmarshal(respB, &parsed); err != nil {
			parsed = string(respB)
		}
	}

	out := map[string]interface{}{
		"statusCode": resp.StatusCode,
		"body":       parsed,
	}
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("ai engine status %d", resp.StatusCode)
	}
	return out, nil
}

func (ne *NodeExecutor) execDecision(ctx context.Context, node models.Node, input map[string]interface{}) (map[string]interface{}, error) {
	// Expect node.Data["condition"] — we'll do a simple evaluation:
	// If condition equals a key in 'input' and the value is truthy -> true.
	// If condition is "yes"/"true"/"1" -> true, else false.
	condRaw, ok := node.Data["condition"]
	cond := ""
	if ok {
		cond = fmt.Sprintf("%v", condRaw)
	}
	cond = strings.TrimSpace(strings.ToLower(cond))

	result := false
	if cond == "true" || cond == "yes" || cond == "1" {
		result = true
	} else if val, ok := input[condRaw.(string)]; ok {
		// if input contains key matching condRaw, evaluate truthiness
		switch v := val.(type) {
		case bool:
			result = v
		case string:
			l := strings.ToLower(strings.TrimSpace(v))
			result = l == "true" || l == "yes" || l == "1"
		case float64:
			result = v != 0
		case int:
			result = v != 0
		default:
			result = v != nil
		}
	}

	return map[string]interface{}{"decision": result}, nil
}
