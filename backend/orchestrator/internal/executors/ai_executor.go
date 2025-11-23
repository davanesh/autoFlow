package executors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Davanesh/auto-orchestrator/internal/services"
)

type AIExecutor struct{}

func (e *AIExecutor) Execute(node *services.ExecNode, g *services.ExecGraph) (string, error) {
	// Take data from the node
	prompt := fmt.Sprintf("%v", node.Data["prompt"])
	input := fmt.Sprintf("%v", node.Data["input"])
	if prompt == "" && input == "" {
		return "", errors.New("AI node requires 'prompt' or 'input'")
	}

	// Build final prompt
	fullPrompt := prompt + "\n\n" + input

	//-------------------------------------------
	// 🔥 OLLAMA REQUEST BODY
	//-------------------------------------------
	reqBody := map[string]interface{}{
		"model":  "llama3.2:1b", // or "llama3", "llama3.1", etc.
		"prompt": fullPrompt,
	}

	jsonBody, _ := json.Marshal(reqBody)

	//-------------------------------------------
	// 🔥 SEND REQUEST TO OLLAMA
	//-------------------------------------------
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

//-------------------------------------------
// 🔥 READ STREAMED RESPONSE (CHUNK BY CHUNK)
//-------------------------------------------
var fullResponse string

decoder := json.NewDecoder(bytes.NewReader(respBytes))

for {
	var chunk struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := decoder.Decode(&chunk); err != nil {
		break // no more chunks
	}

	fullResponse += chunk.Response

	if chunk.Done {
		break
	}
}

if fullResponse == "" {
	return "", errors.New("ollama returned empty response")
}

// SAVE OUTPUT
node.Data["output"] = fullResponse
node.Status = "done"

return "", nil
}

// Register this executor
func init() {
	services.RegisterExecutor("ai", &AIExecutor{})
}
