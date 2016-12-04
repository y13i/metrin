package metrin

import (
	"fmt"

	"reflect"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// CheckThresholds - threshold values and its presense
type CheckThresholds struct {
	CriticalGtPresent  bool
	CriticalLtPresent  bool
	CriticalGtePresent bool
	CriticalLtePresent bool
	WarningGtPresent   bool
	WarningLtPresent   bool
	WarningGtePresent  bool
	WarningLtePresent  bool
	CriticalGtValue    float64
	CriticalLtValue    float64
	CriticalGteValue   float64
	CriticalLteValue   float64
	WarningGtValue     float64
	WarningLtValue     float64
	WarningGteValue    float64
	WarningLteValue    float64
}

// CheckInput - the argument of check
type CheckInput struct {
	Thresholds         CheckThresholds
	Datapoints         []*cloudwatch.Datapoint
	Statistics         []string
	ExtendedStatistics []string
}

// CheckOutput - the result of check, includes message and exit code
type CheckOutput struct {
	ExitCode int
	Messages []string
}

// Check - performs check
func Check(input CheckInput) CheckOutput {
	lastDatapoint := GetLastDatapoint(input.Datapoints)
	lastDatapointValue := getDatapointValue(lastDatapoint, input.Statistics, input.ExtendedStatistics)
	exitCode := 0

	messages := []string{
		fmt.Sprintf("got `%f` (%s)", lastDatapointValue, *lastDatapoint.Unit),
	}

	if input.Thresholds.CriticalGtPresent && lastDatapointValue > input.Thresholds.CriticalGtValue {
		exitCode = 2
		messages = append(messages, fmt.Sprintf("greater than `%f`", input.Thresholds.CriticalGtValue))
	} else if input.Thresholds.CriticalLtPresent && lastDatapointValue < input.Thresholds.CriticalLtValue {
		exitCode = 2
		messages = append(messages, fmt.Sprintf("less than `%f`", input.Thresholds.CriticalLtValue))
	} else if input.Thresholds.CriticalGtePresent && lastDatapointValue >= input.Thresholds.CriticalGteValue {
		exitCode = 2
		messages = append(messages, fmt.Sprintf("greater than or equal to `%f`", input.Thresholds.CriticalGteValue))
	} else if input.Thresholds.CriticalLtePresent && lastDatapointValue <= input.Thresholds.CriticalLteValue {
		exitCode = 2
		messages = append(messages, fmt.Sprintf("less than or equal to `%f`", input.Thresholds.CriticalLteValue))
	} else if input.Thresholds.WarningGtPresent && lastDatapointValue > input.Thresholds.WarningGtValue {
		exitCode = 1
		messages = append(messages, fmt.Sprintf("greater than `%f`", input.Thresholds.WarningGtValue))
	} else if input.Thresholds.WarningLtPresent && lastDatapointValue < input.Thresholds.WarningLtValue {
		exitCode = 1
		messages = append(messages, fmt.Sprintf("less than `%f`", input.Thresholds.WarningLtValue))
	} else if input.Thresholds.WarningGtePresent && lastDatapointValue >= input.Thresholds.WarningGteValue {
		exitCode = 1
		messages = append(messages, fmt.Sprintf("greater than or equal to `%f`", input.Thresholds.WarningGteValue))
	} else if input.Thresholds.WarningLtePresent && lastDatapointValue <= input.Thresholds.WarningLteValue {
		exitCode = 1
		messages = append(messages, fmt.Sprintf("less than or equal to `%f`", input.Thresholds.WarningLteValue))
	}

	switch exitCode {
	case 0:
		messages = append([]string{"CloudWatch OK"}, messages...)
	case 1:
		messages = append([]string{"CloudWatch WARNING"}, messages...)
	case 2:
		messages = append([]string{"CloudWatch CRITICAL"}, messages...)
	}

	return CheckOutput{
		ExitCode: exitCode,
		Messages: messages,
	}
}

func getDatapointValue(datapoint *cloudwatch.Datapoint, statistics []string, extendedStatistics []string) float64 {
	var value float64

	if len(statistics) > 0 {
		value = getStatisticValue(datapoint, statistics[0])
	} else if len(extendedStatistics) > 0 {
		value = getExtendedStatisticValue(datapoint, extendedStatistics[0])
	}

	return value
}

func getStatisticValue(datapoint *cloudwatch.Datapoint, statistic string) float64 {
	r := reflect.Indirect(reflect.ValueOf(datapoint))
	f := r.FieldByName(statistic).Interface().(*float64)

	return float64(*f)
}

func getExtendedStatisticValue(datapoint *cloudwatch.Datapoint, extendedStatistic string) float64 {
	return float64(*datapoint.ExtendedStatistics[extendedStatistic])
}
