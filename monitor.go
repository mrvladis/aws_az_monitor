package main

import (
    "context"
    "fmt"
    "log"
	"time"

    "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/autoscaling"
	asgtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
     cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func main() {
    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatalf("Unable to load SDK config: %v", err)
    }

    // Create  clients
    client := autoscaling.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)
    // Get all Auto Scaling groups
    listASGs, err := getASGs(client)
    if err != nil {
        log.Fatalf("Unable to describe Auto Scaling groups: %v", err)
    }

    // Process each Auto Scaling group
    for _, asg := range listASGs {

		// Create a map to store instance counts per AZ
        azCounts := make(AZInstanceCount)

		// Count instances per AZ that are healthy and in service
        for _, instance := range asg.Instances {
            if isHealthyAndInService(instance) {
                azCounts[*instance.AvailabilityZone]++
            }
        }
		// Print and send metrics to CloudWatch
        fmt.Printf("\nAuto Scaling Group: %s\n", *asg.AutoScalingGroupName)
        fmt.Printf("Healthy and InService instances per AZ:\n")

        if len(azCounts) == 0 {
            fmt.Println("  No healthy and in-service instances found")
        } else {
            // Prepare metrics data
            var metricData []cwtypes.MetricDatum
            totalInstances := 0

            // Create metrics for each AZ
            for az, count := range azCounts {
                fmt.Printf("  %s: %d instances\n", az, count)
                totalInstances += count

                // Create metric for this AZ
                metricData = append(metricData, cwtypes.MetricDatum{
                    MetricName: aws.String("HealthyInstancesInAZ"),
                    Value:      aws.Float64(float64(count)),
                    Timestamp:  aws.Time(time.Now()),
                    Dimensions: []cwtypes.Dimension{
                        {
                            Name:  aws.String("AutoScalingGroupName"),
                            Value: asg.AutoScalingGroupName,
                        },
                        {
                            Name:  aws.String("AvailabilityZone"),
                            Value: aws.String(az),
                        },
                    },
                    Unit: cwtypes.StandardUnitCount,
                })
            }

            // Add total instances metric
            metricData = append(metricData, cwtypes.MetricDatum{
                MetricName: aws.String("TotalHealthyInstances"),
                Value:      aws.Float64(float64(totalInstances)),
                Timestamp:  aws.Time(time.Now()),
                Dimensions: []cwtypes.Dimension{
                    {
                        Name:  aws.String("AutoScalingGroupName"),
                        Value: asg.AutoScalingGroupName,
                    },
                },
                Unit: cwtypes.StandardUnitCount,
            })

            fmt.Printf("Total healthy and in-service instances: %d\n", totalInstances)

            // Send metrics to CloudWatch
            err = sendMetricsToCloudWatch(cwClient, metricData)
            if err != nil {
                log.Printf("Error sending metrics for ASG %s: %v\n", *asg.AutoScalingGroupName, err)
            }
        }
    }
}

func getASGs (client *autoscaling.Client) ([]asgtypes.AutoScalingGroup, error) {
    // Create the input parameters
    input := &autoscaling.DescribeAutoScalingGroupsInput{}

    // Get all Auto Scaling groups
    result, err := client.DescribeAutoScalingGroups(context.TODO(), input)
    if err != nil {
        log.Fatalf("Unable to describe Auto Scaling groups: %v", err)
    }

	return result.AutoScalingGroups, err
}


// Helper function to check if an instance is both healthy and in service
func isHealthyAndInService(instance asgtypes.Instance) bool {
    return instance.HealthStatus != nil &&
           *instance.HealthStatus == "Healthy" &&
           instance.LifecycleState == "InService"
}

func sendMetricsToCloudWatch(client *cloudwatch.Client, metricData []cwtypes.MetricDatum) error {
    // CloudWatch API can only process 20 metrics at a time
    batchSize := 20
    for i := 0; i < len(metricData); i += batchSize {
        end := i + batchSize
        if end > len(metricData) {
            end = len(metricData)
        }

        input := &cloudwatch.PutMetricDataInput{
            Namespace:  aws.String("CustomASGMetrics"),
            MetricData: metricData[i:end],
        }

        _, err := client.PutMetricData(context.TODO(), input)
        if err != nil {
            return fmt.Errorf("error putting metric data: %v", err)
        }
    }
    return nil
}
