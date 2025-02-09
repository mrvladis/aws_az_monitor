# AWS Auto Scaling Group Monitor

A Go application that monitors AWS Auto Scaling Groups (ASGs) and reports instance health metrics to CloudWatch.

## Description

This application provides real-time monitoring of AWS Auto Scaling Groups by:
- Tracking the number of healthy and in-service instances per Availability Zone
- Reporting metrics to CloudWatch for monitoring and alerting
- Providing console output of instance distribution across AZs

## Features

- Retrieves information about all Auto Scaling Groups in your AWS account
- Counts instances by state (running, pending, terminated, etc.) per Availability Zone
- Publishes detailed instance state metrics to CloudWatch
- Maintains backwards compatibility with existing healthy instance metrics
- Handles CloudWatch API limitations with batch processing
- Provides detailed console output of instance distribution across AZs by state
- Monitors Amazon RDS clusters for:
  - Writer and reader instance distribution across Availability Zones
  - Real-time instance role tracking (writer/reader)


## Prerequisites

- Go 1.x or higher
- AWS credentials configured
- Appropriate IAM permissions:
  - `autoscaling:DescribeAutoScalingGroups`
  - `cloudwatch:PutMetricData`
  - `rds:DescribeDBClusters`
  - `rds:DescribeDBInstances`


## Installation

1. Clone the repository
2. Install dependencies:
```bash
go mod tidy

```

# Configuration

Ensure your AWS credentials are properly configured using one of the following methods:

- AWS CLI configuration
- Environment variables
- IAM role (if running on AWS infrastructure)

# Usage

## Running

Run the application:

```bash
go run monitor.go
```
## Building

### Binary Builds
You can build the binary for different architectures using Go's cross-compilation capabilities:

```bash
# Build for Linux AMD64
GOOS=linux GOARCH=amd64 go build -o monitor-linux-amd64

# Build for Linux ARM64
GOOS=linux GOARCH=arm64 go build -o monitor-linux-arm64

# Build for macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o monitor-darwin-amd64

# Build for macOS ARM64 (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o monitor-darwin-arm64

# Build for Windows AMD64
GOOS=windows GOARCH=amd64 go build -o monitor-windows-amd64.exe
```

### Container Build
#### Dockerfile
Create a Dockerfile in your project root:

```bash
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /monitor

# Final stage
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /monitor .

# Add CA certificates for AWS API calls
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D appuser
USER appuser

ENTRYPOINT ["./monitor"]
```
#### Build the container
docker build -t asg-monitor:latest .

##### Run the container with AWS credentials
docker run -d \
    -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e AWS_REGION=${AWS_REGION} \
    asg-monitor:latest

##### If using AWS IAM roles, you can pass the credentials from the host:
docker run -d \
    -v ~/.aws:/home/appuser/.aws:ro \
    asg-monitor:latest

# CloudWatch Metrics

The application publishes metrics under two namespaces:

## Auto Scaling Group Metrics

### Instance State Metrics per AZ
The application tracks instances by their state in each Availability Zone:
- Dimensions:
  - AutoScalingGroupName
  - AvailabilityZone
- Metrics are created for each EC2 instance state (e.g., "running", "pending", "terminated", etc.)
- Unit: Count

### Legacy Metrics (maintained for backwards compatibility)
#### HealthyInstancesInAZ
- Dimensions:
  - AutoScalingGroupName
  - AvailabilityZone
- Unit: Count

#### TotalHealthyInstances
- Dimensions:
  - AutoScalingGroupName
- Unit: Count

### RDS Cluster Metrics
The application monitors RDS clusters and publishes the following metrics:

#### Instance Distribution Metrics per AZ
- Dimensions:
  - DBClusterIdentifier
  - AvailabilityZone
- Metrics:
  - `WriterInstancesInAZ`: Number of writer instances in each AZ
  - `ReaderInstancesInAZ`: Number of reader instances in each AZ
- Unit: Count

The application provides console output showing:
- RDS Cluster ID
- Distribution of writer and reader instances across Availability Zones


# Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

# License

MIT License

Copyright (c) 2024 Vladislav Nedosekin

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```