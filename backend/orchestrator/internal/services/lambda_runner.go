package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaInvoker holds the AWS Lambda client
type LambdaInvoker struct {
	client *lambda.Client
}

// NewLambdaInvoker initializes and returns a LambdaInvoker.
// It loads credentials from the environment/`~/.aws/credentials` (aws configure).
func NewLambdaInvoker() (*LambdaInvoker, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &LambdaInvoker{
		client: lambda.NewFromConfig(cfg),
	}, nil
}

// Invoke invokes a Lambda function by name with a payload (any serializable object).
// Returns the raw payload bytes and the function error (if any).
func (li *LambdaInvoker) Invoke(functionName string, payload interface{}) ([]byte, error) {
	// Marshal the payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	input := &lambda.InvokeInput{
		FunctionName: &functionName,
		Payload:      body,
		// You can set InvocationType: types.InvocationTypeRequestResponse for sync (default)
		InvocationType: types.InvocationTypeRequestResponse,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	out, err := li.client.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("lambda invoke error: %w", err)
	}

	// If the Lambda returned an error (FunctionError set), include that in returned error
	if out.FunctionError != nil && *out.FunctionError != "" {
		return out.Payload, fmt.Errorf("lambda function error: %s", *out.FunctionError)
	}

	return out.Payload, nil
}
