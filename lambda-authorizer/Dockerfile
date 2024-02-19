# Start from the latest golang base image
FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

######## Start a new stage from AWS Lambda Go base image #######
FROM public.ecr.aws/lambda/go:latest

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main /var/task/

# Command can be the function handler
CMD ["./main"]