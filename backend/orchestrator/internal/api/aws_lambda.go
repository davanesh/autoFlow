package api

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type LambdaClient struct {
	client *lambda.Client
}

func NewLambdaClient() (*LambdaClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &LambdaClient{
		client: lambda.NewFromConfig(cfg),
	}, nil
}

func (l *LambdaClient) InvokeLambda(functionName string, payload interface{}) ([]byte, error) {
	body, _ := json.Marshal(payload)

	input := &lambda.InvokeInput{
		FunctionName: &functionName,
		Payload:      body,
	}

	result, err := l.client.Invoke(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return result.Payload, nil
}
