package metrin

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

const (
	defaultRegion = "us-east-1"
)

// BuildParamsInput - the type for BuildParams function argument
type BuildParamsInput struct {
	Namespace          string
	MetricName         string
	StartTime          int64
	EndTime            int64
	Period             int64
	Unit               string
	Statistics         []string
	ExtendedStatistics []string
	Dimensions         []string
}

// BuildParams - used for create GetMetricStatisticsInput
func BuildParams(input BuildParamsInput) *cloudwatch.GetMetricStatisticsInput {
	var startTime time.Time
	var endTime time.Time

	if input.StartTime <= 0 {
		startTime = time.Unix((time.Now().Unix() + input.StartTime), 0)
	}

	if input.EndTime <= 0 {
		endTime = time.Unix((time.Now().Unix() + input.EndTime), 0)
	}

	params := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(input.Namespace),
		MetricName: aws.String(input.MetricName),
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int64(input.Period),
	}

	if len(input.Unit) > 0 {
		params.Unit = aws.String(input.Unit)
	}

	if len(input.Statistics) > 0 {
		s := make([]*string, len(input.Statistics))

		for i := range input.Statistics {
			s[i] = aws.String(input.Statistics[i])
		}

		params.Statistics = s
	}

	if len(input.ExtendedStatistics) > 0 {
		s := make([]*string, len(input.ExtendedStatistics))

		for i := range input.ExtendedStatistics {
			s[i] = aws.String(input.ExtendedStatistics[i])
		}

		params.ExtendedStatistics = s
	}

	if len(input.Dimensions) > 0 {
		s := make([]*cloudwatch.Dimension, len(input.Dimensions))

		for i := range input.Dimensions {
			splitted := strings.Split(input.Dimensions[i], ":")

			s[i] = &cloudwatch.Dimension{
				Name:  aws.String(splitted[0]),
				Value: aws.String(splitted[1]),
			}
		}

		params.Dimensions = s
	}

	return params
}

// GetMetricStatistics wrapper
func GetMetricStatistics(params *cloudwatch.GetMetricStatisticsInput) *cloudwatch.GetMetricStatisticsOutput {
	fmt.Println(os.Getenv("AWS_REGION"))
	if os.Getenv("AWS_REGION") == "" {
		os.Setenv("AWS_REGION", defaultRegion)
	}

	session, err := session.NewSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	service := cloudwatch.New(session)

	response, err := service.GetMetricStatistics(params)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return response
}

// GetLastDatapoint - get latest datapoint from datapoints by timestamp
func GetLastDatapoint(datapoints []*cloudwatch.Datapoint) *cloudwatch.Datapoint {
	lastDatapoint := datapoints[0]

	for i := range datapoints {
		datapoint := datapoints[i]

		if datapoint.Timestamp.Unix() > lastDatapoint.Timestamp.Unix() {
			lastDatapoint = datapoint
		}
	}

	return lastDatapoint
}
