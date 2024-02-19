package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"strings"
)

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	tokenString := event.AuthorizationToken
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	log.Println("Method ARN: " + event.MethodArn)
	log.Println("Token: " + tokenString)

	tokenParts := strings.Split(tokenString, ".")
	if len(tokenParts) != 3 {
		return generatePolicy("user", "Deny", event.MethodArn), nil
	}

	return generatePolicy("user", "Allow", event.MethodArn), nil
}

func generatePolicy(principalID, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalID}
	if resource == "" {
		resource = "*"
	}
	authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action:   []string{"execute-api:Invoke"},
				Effect:   effect,
				Resource: []string{resource},
			},
		},
	}
	return authResponse
}

func main() {
	lambda.Start(handler)
}
