# Lambda Web Service Demo 

This project demonstrates the use of AWS Lambda functions to create a web service with token validation, HTTP response handling, and canary deployment process.

## Overview

The system consists of three main components:

- `lambda-authorizer`: A Lambda function responsible for token validation, ensuring that only valid requests are processed by the service.
- `lambda-code`: A Lambda function that provides example HTTP responses. This could represent the core functionality of your web service.
- `lambda-deployer`: A Lambda function that manages the canary deployment of `lambda-code`. Canary deployment allows new versions to be released to a subset of users before a full rollout.

## Canary Deployment

The canary deployment is orchestrated by the `lambda-deployer`, which is triggered by messages on an AWS SQS queue. When a new deployment is initiated, `lambda-deployer` gradually shifts traffic to the new version of `lambda-code`, monitors its health, and finalizes or rolls back the deployment based on certain health checks.

Depending on deployment time and traffic, the canary deployment could take more than 15 minutes to complete which is the limit of the AWS Lambda function execution time. In that case, we can use CodeDeploy to implement the canary deployment process.

## Infrastructure 

The project uses Terraform to define and manage the AWS infrastructure, ensuring that the environment is reproducible and maintainable. The code includes definitions for the necessary AWS resources such as Lambda functions, IAM roles, API Gateway, CloudWatch, VPC, and SQS.

## Getting Started

To get started with this project, clone the repository to your local machine and ensure you have the following prerequisites installed:

- AWS CLI
- Terraform

Configure your AWS CLI with the appropriate credentials and run the following commands to initialize your Terraform workspace:

```sh
terraform init
terraform apply
```
