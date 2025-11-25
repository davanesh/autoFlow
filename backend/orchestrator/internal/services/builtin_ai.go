package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AIExecutor struct{}

func init() {
	RegisterExecutor("ai", &AIExecutor{})
}

func (e *AIExecutor) Execute(node *ExecNode, g *ExecGraph) (string, error) {
	prompt := fmt.Sprintf("%v", node.Data["prompt"])
	input := fmt.Sprintf("%v", node.Data["input"])

	if prompt == "" && input == "" {
		return "", errors.New("AI node requires 'prompt' or 'input'")
	}

	fullPrompt := prompt + "\n\n" + input

	reqBody := map[string]interface{}{
		"model":  "llama3.2:1b",
		"prompt": fullPrompt,
	}

	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	log.Println("🔍 RAW OLLAMA RESPONSE:", string(respBytes))

	var fullResponse string
	decoder := json.NewDecoder(bytes.NewReader(respBytes))

	for {
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}

		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		fullResponse += chunk.Response
		if chunk.Done {
			break
		}
	}

	if fullResponse == "" {
		return "", errors.New("ollama returned empty response")
	}

	node.Data["output"] = fullResponse
	node.Status = "done"
	return "", nil
}
