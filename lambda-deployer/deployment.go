package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/lambda"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Deployment struct {
	event   *DeploymentEvent
	session *session.Session
}

func (d *Deployment) deploy() error {
	// Calculate initial weights for traffic shifting
	currentWeight := d.event.InitialWeight
	maxWeight := 100
	increment := d.event.Increment
	checkInterval := time.Duration(d.event.CheckInterval) * time.Second

	functionConfigs, err := d.updateLambdaFunction()
	if err != nil {
		fmt.Printf("Error updating Lambda function: %s\n", err)
		return fmt.Errorf("failed to update Lambda function: %w", err)
	}

	for currentWeight < maxWeight {
		// Step 2: Update the weight
		if err := d.updateLambdaAlias(currentWeight, *functionConfigs.Version); err != nil {
			return fmt.Errorf("failed to update alias weight: %w", err)
		}

		// Step 3: Wait for the specified interval
		time.Sleep(checkInterval)

		// Step 4 & 5: Perform and verify health check
		if !d.performHealthCheck() {
			// If the health check fails, rollback
			if rollbackErr := d.rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to rollback after unsuccessful health check: %w", rollbackErr)
			}
			return fmt.Errorf("deployment failed after unsuccessful health check")
		}

		// If the health check is successful and we've reached full traffic, finalize
		if currentWeight == maxWeight {
			break // Exit loop to finalize
		}

		// Increment the weight for the next iteration
		currentWeight += increment
		if currentWeight > maxWeight {
			currentWeight = maxWeight // Ensure we don't exceed the maximum
		}
	}

	// Step 6: Finalize the deployment
	if err := d.finalize(*functionConfigs.Version); err != nil {
		return fmt.Errorf("failed to finalize deployment: %w", err)
	}

	return nil // Deployment successful
}

func (d *Deployment) performHealthCheck() bool {
	cw := cloudwatch.New(d.session)

	funcPlusAlias := d.event.FunctionName + ":" + d.event.AliasName
	fmt.Printf("Checking health of new version: %s\n", funcPlusAlias)
	now := time.Now().UTC()
	startTime := now.Add(-1 * time.Minute)

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/Lambda"),
		MetricName: aws.String("Errors"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("FunctionName"),
				Value: aws.String(d.event.FunctionName),
			},
			{
				Name:  aws.String("Resource"),
				Value: aws.String(funcPlusAlias),
			},
			{
				Name:  aws.String("ExecutedVersion"),
				Value: aws.String(d.event.AliasName),
			},
		},
		StartTime: aws.Time(startTime),
		EndTime:   aws.Time(now),
		Period:    aws.Int64(60),
		Statistics: []*string{
			aws.String("Sum"),
		},
	}

	result, err := cw.GetMetricStatistics(input)
	if err != nil {
		fmt.Println("Error getting metric statistics:", err)
		return false
	}

	for _, datapoint := range result.Datapoints {
		if *datapoint.Sum > 0 {
			fmt.Printf("Failing health check because error metrics were found for new version: %v\n", datapoint)
			return false
		}
	}

	return true
}

func (d *Deployment) updateLambdaAlias(nextWeight int, version string) error {
	fmt.Printf("Next weight: %d\n", nextWeight)

	// Create Lambda service client
	svc := lambda.New(d.session)

	routingConfig := &lambda.AliasRoutingConfiguration{
		AdditionalVersionWeights: map[string]*float64{
			version: aws.Float64(float64(nextWeight) / 100), // Assuming nextWeight is a percentage
		},
	}

	// The 'Name' field should be just the alias name, not a combination of function name and alias name
	input := &lambda.UpdateAliasInput{
		FunctionName:  aws.String(d.event.FunctionName),
		Name:          aws.String(d.event.AliasName), // Use just the alias name here
		RoutingConfig: routingConfig,
	}

	result, err := svc.UpdateAlias(input)
	if err != nil {
		fmt.Println("Error updating alias:", err)
		return err
	}

	fmt.Println("Update alias result:", result)
	return nil
}

// updateLambdaFunction updates the Lambda function's code with a new container image URL and publishes a new version.
func (d *Deployment) updateLambdaFunction() (*lambda.FunctionConfiguration, error) {
	svc := lambda.New(d.session)

	// Step 1: Update the Lambda function code with the new image URL
	updateCodeInput := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(d.event.FunctionName),
		ImageUri:     aws.String(d.event.ImageURL),
	}

	_, err := svc.UpdateFunctionCode(updateCodeInput)
	if err != nil {
		fmt.Printf("Error updating function code: %s\n", err)
		return nil, err
	}

	for {
		getFunctionInput := &lambda.GetFunctionInput{
			FunctionName: aws.String(d.event.FunctionName),
		}
		function, err := svc.GetFunction(getFunctionInput)
		if err != nil {
			fmt.Printf("Error getting function: %s\n", err)
			return nil, err
		}

		state := *function.Configuration.State
		lastUpdateStatus := *function.Configuration.LastUpdateStatus

		fmt.Printf("Function state: %s, Last update status: %s\n", state, lastUpdateStatus)

		// Check for successful update
		if state == "Active" && lastUpdateStatus == "Successful" {
			break // Update successful, exit the loop
		}

		// Exit early if the update has failed
		if lastUpdateStatus == "Failed" {
			fmt.Printf("Update failed for function: %s\n", d.event.FunctionName)
			return nil, fmt.Errorf("update failed for function: %s, state: %s", d.event.FunctionName, state)
		}

		// If the function is still updating, wait before checking again
		if lastUpdateStatus == "InProgress" {
			fmt.Println("Update in progress, waiting for function to become active...")
			time.Sleep(2 * time.Second) // Adjust the sleep time as needed
			continue
		}

		// Handle unexpected statuses
		fmt.Printf("Unexpected function update status: %s\n", lastUpdateStatus)
		return nil, fmt.Errorf("unexpected update status for function: %s, status: %s", d.event.FunctionName, lastUpdateStatus)
	}

	// Step 2: Publish a new version of the Lambda function
	publishVersionInput := &lambda.PublishVersionInput{
		FunctionName: aws.String(d.event.FunctionName),
	}

	// The response includes the new version's configuration
	newVersion, err := svc.PublishVersion(publishVersionInput)
	if err != nil {
		fmt.Printf("Error publishing new version: %s\n", err)
		return nil, err
	}

	fmt.Printf("Published new version: %s\n", *newVersion.Version)
	return newVersion, nil
}

func (d *Deployment) rollback() error {
	fmt.Println("Rolling back: removing traffic shifting from alias")

	// Create Lambda service client
	svc := lambda.New(d.session)

	// Update alias to remove traffic shifting by setting AdditionalVersionWeights to an empty map
	input := &lambda.UpdateAliasInput{
		FunctionName: aws.String(d.event.FunctionName),
		Name:         aws.String(d.event.AliasName),
		RoutingConfig: &lambda.AliasRoutingConfiguration{
			AdditionalVersionWeights: make(map[string]*float64), // Empty map removes traffic shifting
		},
	}

	_, err := svc.UpdateAlias(input)
	if err != nil {
		fmt.Printf("Error rolling back alias: %s\n", err)
		return err
	}

	fmt.Println("Rollback successful: traffic shifting removed")
	return nil
}

func (d *Deployment) finalize(newVersion string) error {
	fmt.Printf("Finalizing update: setting alias to point to version %s\n", newVersion)

	// Create Lambda service client
	svc := lambda.New(d.session)

	// Update alias to point directly to the new version and remove any traffic shifting
	input := &lambda.UpdateAliasInput{
		FunctionName:    aws.String(d.event.FunctionName), // Name of the Lambda function
		Name:            aws.String(d.event.AliasName),    // Name of the alias
		FunctionVersion: aws.String(newVersion),           // The new version to point the alias to
		RoutingConfig: &lambda.AliasRoutingConfiguration{
			AdditionalVersionWeights: make(map[string]*float64), // Ensure routing config is empty
		},
	}

	_, err := svc.UpdateAlias(input)
	if err != nil {
		fmt.Printf("Error finalizing alias update: %s\n", err)
		return err
	}

	fmt.Println("Finalize successful: Alias updated to new version")
	return nil
}

func (d *Deployment) rollbackOnInterrupt() {
	// Create a channel to receive signal notifications.
	sigChan := make(chan os.Signal, 1)

	// Register the channel to receive notifications for SIGTERM and SIGINT.
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// This goroutine executes a blocking receive for signals.
	// When it gets one it will perform the rollback.
	go func() {
		// Block until a signal is received.
		sig := <-sigChan
		fmt.Printf("Caught signal %s: starting rollback.\n", sig)

		// Perform rollback.
		err := d.rollback()
		if err != nil {
			fmt.Printf("Error during rollback: %v\n", err)
			os.Exit(1) // Exit with error status.
		}

		fmt.Println("Rollback completed successfully.")
		os.Exit(0) // Exit normally.
	}()
}
