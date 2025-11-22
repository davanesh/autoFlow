package executors

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/Davanesh/auto-orchestrator/internal/services"
)

/*
------------------------------------------
    AI EXECUTOR (OpenAI API)
------------------------------------------
*/

type AiExecutor struct {
	httpClient *http.Client
}

func NewAiExecutor() *AiExecutor {
	return &AiExecutor{
		httpClient: &http.Client{},
	}
}

// OpenAI request structure
type openAIChatReq struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI response structure
type openAIChatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (e *AiExecutor) Execute(node *services.ExecNode, g *services.ExecGraph) (string, error) {

	// Get AI prompt
	prompt, ok := node.Data["prompt"].(string)
	if !ok || prompt == "" {
		return "", errors.New("AI node requires a 'prompt' field")
	}

	// Get user input
	input, _ := node.Data["input"].(string)

	// Fetch the OpenAI key from env
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing OPENAI_API_KEY (add it to .env)")
	}

	log.Println("🤖 Running AI Node:", node.Label)

	reqBody := openAIChatReq{
		Model: "gpt-4o-mini", // cheap + fast model
		Messages: []chatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: input},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var aiResp openAIChatResp
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", errors.New("AI returned no message")
	}

	output := aiResp.Choices[0].Message.Content

	// Save result for next nodes
	node.Data["output"] = output
	node.Status = "done"

	log.Println("🤖 AI Output:", output)

	// Move to next node
	if len(node.Next) > 0 {
		return node.Next[0], nil
	}

	return "", nil
}

/*
-----------------------------------
   Register AI executor globally
-----------------------------------
*/

func init() {
  services.RegisterExecutor("ai", NewAiExecutor())
}
