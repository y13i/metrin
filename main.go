package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/y13i/metrin/lib"

	"github.com/urfave/cli"
)

const (
	defaultStartTime = -300
	defaultEndTime   = 0
	defaultPeriod    = 60
)

func main() {
	app := cli.NewApp()

	app.Name = "metrin"
	app.Usage = "Very simple CloudWatch CLI for Zabbix/Nagios/Sensu/Mackerel/etc."
	app.Version = "0.0.4"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "namespace, n",
			Usage: "CloudWatch namespace. e.g. 'AWS/EC2'",
		},

		cli.StringFlag{
			Name:  "metric-name, m",
			Usage: "CloudWatch metric name. e.g. 'CPUUtilization'",
		},

		cli.Int64Flag{
			Name:  "start-time, S",
			Value: defaultStartTime,
			Usage: "start time as unix timestamp, relative from now if 0 or negative value given",
		},

		cli.Int64Flag{
			Name:  "end-time, E",
			Value: defaultEndTime,
			Usage: "end time as unix timestamp, relative from now if 0 or negative value given",
		},

		cli.Int64Flag{
			Name:  "period, p",
			Usage: "CloudWatch metric statistic period.",
			Value: defaultPeriod,
		},

		cli.StringFlag{
			Name:  "unit, u",
			Usage: "CloudWatch metric statistic unit. e.g. 'Percent'",
		},

		cli.StringSliceFlag{
			Name:  "statistic, s",
			Usage: "CloudWatch metrics statistic. e.g. 'Average'",
		},

		cli.StringSliceFlag{
			Name:  "extended-statistic, e",
			Usage: "CloudWatch extended metrics statistic. e.g. 'p99.5'",
		},

		cli.StringSliceFlag{
			Name:  "dimension, d",
			Usage: "CloudWatch dimension. `DIM_KEY:DIM_VALUE` e.g. 'InstanceId:i-12345678'",
		},

		cli.StringFlag{
			Name:  "region, r",
			Usage: "AWS region. e.g. 'us-west-2'",
		},

		cli.StringFlag{
			Name:  "profile, P",
			Usage: "AWS profile name. e.g. 'myprofile'",
		},

		cli.StringFlag{
			Name:  "access-key-id, a",
			Usage: "AWS access key id. e.g. 'AKIAIOSFODNN7EXAMPLE'",
		},

		cli.StringFlag{
			Name:  "secret-access-key, A",
			Usage: "AWS secret access key. e.g. 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "check",
			Usage: "perform check and exit with status codes (0: OK, 1: WARNING, 2: CRITICAL, 3: UNKNOWN)",

			Flags: []cli.Flag{
				cli.Float64Flag{
					Name:  "critical-gt",
					Usage: "exit as critical (code 2) if latest metric value is greater than the value",
				},

				cli.Float64Flag{
					Name:  "critical-lt",
					Usage: "exit as critical (code 2) if latest metric value is less than the value",
				},

				cli.Float64Flag{
					Name:  "critical-gte",
					Usage: "exit as critical (code 2) if latest metric value is greater than or equal to the value",
				},

				cli.Float64Flag{
					Name:  "critical-lte",
					Usage: "exit as critical (code 2) if latest metric value is less than or equal to the value",
				},

				cli.Float64Flag{
					Name:  "warning-gt",
					Usage: "exit as warning (code 1) if latest metric value is greater than the value",
				},

				cli.Float64Flag{
					Name:  "warning-lt",
					Usage: "exit as warning (code 1) if latest metric value is less than the value",
				},

				cli.Float64Flag{
					Name:  "warning-gte",
					Usage: "exit as warning (code 1) if latest metric value is greater than or equal to the value",
				},

				cli.Float64Flag{
					Name:  "warning-lte",
					Usage: "exit as warning (code 1) if latest metric value is less than or equal to the value",
				},
			},

			Action: func(ctx *cli.Context) error {
				setAwsEnv(ctx)
				params := getParams(ctx)
				response := metrin.GetMetricStatistics(params)

				thresholds := metrin.CheckThresholds{
					CriticalGtPresent:  ctx.IsSet("critical-gt"),
					CriticalLtPresent:  ctx.IsSet("critical-lt"),
					CriticalGtePresent: ctx.IsSet("critical-gte"),
					CriticalLtePresent: ctx.IsSet("critical-lte"),
					WarningGtPresent:   ctx.IsSet("warning-gt"),
					WarningLtPresent:   ctx.IsSet("warning-lt"),
					WarningGtePresent:  ctx.IsSet("warning-gte"),
					WarningLtePresent:  ctx.IsSet("warning-lte"),
					CriticalGtValue:    ctx.Float64("critical-gt"),
					CriticalLtValue:    ctx.Float64("critical-lt"),
					CriticalGteValue:   ctx.Float64("critical-gte"),
					CriticalLteValue:   ctx.Float64("critical-lte"),
					WarningGtValue:     ctx.Float64("warning-gt"),
					WarningLtValue:     ctx.Float64("warning-lt"),
					WarningGteValue:    ctx.Float64("warning-gte"),
					WarningLteValue:    ctx.Float64("warning-lte"),
				}

				checkOutput := metrin.Check(metrin.CheckInput{
					Thresholds:         thresholds,
					Datapoints:         response.Datapoints,
					Statistics:         ctx.GlobalStringSlice("statistic"),
					ExtendedStatistics: ctx.GlobalStringSlice("extended-statistic"),
				})

				fmt.Println(strings.Join(checkOutput.Messages, ", "))
				fmt.Println("Params:", params)
				os.Exit(checkOutput.ExitCode)

				return nil
			},
		},

		{
			Name:  "print",
			Usage: "Prints GetMetricStatistics response with given format template",

			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "template, t",
					Usage: "output format template (using 'text/template' package. see https://golang.org/pkg/text/template/)",
					Value: "CloudWatch.{{(index .Params.Dimensions 0).Name}}.{{(index .Params.Dimensions 0).Value}}.{{.Params.MetricName}}.{{index .Params.Statistics 0}}\t{{getvalue .Datapoint .Params 0 | deref | printf \"%f\"}}\t{{.Datapoint.Timestamp | unixtime}}",
				},

				cli.BoolFlag{
					Name:  "last-value-only",
					Usage: "if true, print last datapoint value only",
				},
			},

			Action: func(ctx *cli.Context) error {
				setAwsEnv(ctx)
				params := getParams(ctx)
				response := metrin.GetMetricStatistics(params)

				var datapoints []*cloudwatch.Datapoint

				if ctx.Bool("last-value-only") {
					datapoints = []*cloudwatch.Datapoint{
						metrin.GetLastDatapoint(response.Datapoints),
					}
				} else {
					datapoints = response.Datapoints
				}

				outputStrings := metrin.BuildPrintStrings(metrin.BuildPrintStringInput{
					Params:         params,
					Datapoints:     datapoints,
					TemplateString: ctx.String("template"),
				})

				fmt.Println(strings.Join(outputStrings, "\n"))

				return nil
			},
		},

		{
			Name:  "debug",
			Usage: "Prints GetMetricStatistics params and response",

			Action: func(ctx *cli.Context) error {
				setAwsEnv(ctx)
				params := getParams(ctx)
				fmt.Println("Params:", params)

				response := metrin.GetMetricStatistics(params)
				fmt.Println("Response:", response)

				return nil
			},
		},
	}

	app.Run(os.Args)
}

func getParams(ctx *cli.Context) *cloudwatch.GetMetricStatisticsInput {
	return metrin.BuildParams(metrin.BuildParamsInput{
		Namespace:          ctx.GlobalString("namespace"),
		MetricName:         ctx.GlobalString("metric-name"),
		StartTime:          ctx.GlobalInt64("start-time"),
		EndTime:            ctx.GlobalInt64("end-time"),
		Period:             ctx.GlobalInt64("period"),
		Unit:               ctx.GlobalString("unit"),
		Statistics:         ctx.GlobalStringSlice("statistic"),
		ExtendedStatistics: ctx.GlobalStringSlice("extended-statistic"),
		Dimensions:         ctx.GlobalStringSlice("dimension"),
	})
}

func setAwsEnv(ctx *cli.Context) {
	if ctx.GlobalIsSet("region") {
		os.Setenv("AWS_REGION", ctx.GlobalString("region"))
	}

	if ctx.GlobalIsSet("profile") {
		os.Setenv("AWS_PROFILE", ctx.GlobalString("profile"))
	}

	if ctx.GlobalIsSet("access-key-id") {
		os.Setenv("AWS_ACCESS_KEY_ID", ctx.GlobalString("access-key-id"))
	}

	if ctx.GlobalIsSet("secret-access-key") {
		os.Setenv("AWS_SECRET_ACCESS_KEY", ctx.GlobalString("secret-access-key"))
	}
}
