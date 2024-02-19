package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type DeploymentEvent struct {
	ImageURL      string `json:"image_url"`
	FunctionName  string `json:"function_name"`
	AliasName     string `json:"alias_name"`
	InitialWeight int    `json:"initial_weight"`
	CheckInterval int    `json:"check_interval"`
	Increment     int    `json:"increment"`
}

func main() {
	//lambda.Start(handler)

	// deploymentEvent := &DeploymentEvent{
	// 	ImageURL:      "072422391281.dkr.ecr.us-east-1.amazonaws.com/lambda:v2",
	// 	FunctionName:  "arn:aws:lambda:us-east-1:072422391281:function:api_handler_function",
	// 	AliasName:     "production",
	// 	InitialWeight: 10,
	// 	CheckInterval: 10,
	// 	Increment:     10,
	// }

	//awsProfile := os.Getenv("AWS_PROFILE")
	//sess := session.Must(session.NewSessionWithOptions(session.Options{
	//	Config: aws.Config{
	//		Region:                        aws.String("us-east-1"),
	//		CredentialsChainVerboseErrors: aws.Bool(true),
	//	},
	//	Profile: awsProfile,
	//}))
	//
	//deployment := &Deployment{
	//	event:   deploymentEvent,
	//	session: sess,
	//}
	//
	//err := deployment.deploy()
	//if err != nil {
	//	fmt.Printf("Error deploying: %s\n", err)
	//} else {
	//	fmt.Println("Deployment successful")
	//}

	lambda.Start(handler)
}

func handler(ctx context.Context, event events.SQSEvent) error {
	for _, message := range event.Records {
		deploymentEvent := DeploymentEvent{}
		// Assuming the message body is a JSON string that directly maps to the DeploymentEvent struct
		err := json.Unmarshal([]byte(message.Body), &deploymentEvent)
		if err != nil {
			return fmt.Errorf("error unmarshalling SQS message: %v", err)
		}

		if deploymentEvent.ImageURL == "" || deploymentEvent.FunctionName == "" || deploymentEvent.AliasName == "" || deploymentEvent.Increment == 0 {
			return fmt.Errorf("invalid deployment event: %v", deploymentEvent)
		}

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:                        aws.String("us-east-1"),
				CredentialsChainVerboseErrors: aws.Bool(true),
			},
		}))

		// Continue with the deployment process using the values from the deploymentEvent
		deployment := &Deployment{
			event:   &deploymentEvent,
			session: sess,
		}

		// add a signal handler to rollback the deployment if the process is interrupted
		deployment.rollbackOnInterrupt()

		if err := deployment.deploy(); err != nil {
			return fmt.Errorf("deployment error: %v", err)
		}
	}

	return nil
}
