package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type Response events.APIGatewayProxyResponse

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:                        aws.String("us-east-1"),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	}))

	svc := cloudwatch.New(sess)

	_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("lambda-code"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("requestCount"),
				Unit:       aws.String("Count"),
				Value:      aws.Float64(1.0),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("type"),
						Value: aws.String("poc"),
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println("Error adding metrics:", err)
	} else {
		fmt.Println("Metric added successfully")
	}
	return Response{
		StatusCode: 200,
		Body:       fmt.Sprintf("Hello, world! You are using version v3. Your request was %s", request.Body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
