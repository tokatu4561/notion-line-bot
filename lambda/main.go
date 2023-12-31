package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Name string `json:"name"`
}

// 仮のlambda関数
func handler(ctx context.Context, event MyEvent) (string, error) {
	return "Hello " + event.Name, nil
}

func main() {
	lambda.Start(handler)
}