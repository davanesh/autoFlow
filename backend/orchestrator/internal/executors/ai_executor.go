package executors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Davanesh/auto-orchestrator/internal/services"
)

type AIExecutor struct{}

func (e *AIExecutor) Execute(node *services.ExecNode, g *services.ExecGraph) (string, error) {
	prompt := fmt.Sprintf("%v", node.Data["prompt"])
	input := fmt.Sprintf("%v", node.Data["input"])

	if prompt == "" && input == "" {
		return "", errors.New("AI node requires 'prompt' or 'input'")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing OPENAI_API_KEY")
	}

	// NEW OPENAI RESPONSES API FORMAT (2025)
	body := map[string]interface{}{
		"model": "gpt-4.1-mini",
		"input": []map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": prompt},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": input},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// READ RESPONSE
	respBytes, _ := io.ReadAll(resp.Body)
	log.Println("🔍 RAW OpenAI Response:", string(respBytes))

	var out struct {
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	json.Unmarshal(respBytes, &out)

	// Extract output_text
	var final string
	if len(out.Output) > 0 && len(out.Output[0].Content) > 0 {
		for _, c := range out.Output[0].Content {
			if c.Type == "output_text" {
				final += c.Text
			}
		}
	}

	if final == "" {
		return "", errors.New("OpenAI returned no output_text")
	}

	// SAVE OUTPUT INTO WORKFLOW NODE
	node.Data["output"] = final
	node.Status = "done"

	return "", nil
}


// IMPORTANT: Without this your AI node NEVER runs
func init() {
	fmt.Println("🔥 AI Executor Registered!")
	services.RegisterExecutor("ai", &AIExecutor{})
}

