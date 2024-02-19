package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	return Response{
		StatusCode: 200,
		Body:       fmt.Sprintf("Hello, world! You are using version v3. Your request was %s", request.Body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
