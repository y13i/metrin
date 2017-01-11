# metrin

Very simple CloudWatch CLI for Zabbix/Sensu/Mackerel/etc.

## Installation

Download binary from [releases](https://github.com/y13i/metrin/releases).

Put it into your `$PATH`.

## Usage

View it first.

```
$ metrin --help
```

### Set credentials and region

Use environmental variables. `AWS_REGION`, `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`

### Subcommands

#### `check`

Act as Nagios/Sensu plugin style.

Example: Check an EC2 instance's CPU, critical if > 50%

```
$ metrin --namespace AWS/EC2 --metric-name CPUUtilization --statistic Average --dimension InstanceId:i-1234abcd1234abcde check --critical-gt 50
CloudWatch OK, got `22.534000` (Percent)
Params: {
  Dimensions: [{
      Name: "InstanceId",
      Value: "i-1234abcd1234abcde"
    }],
  EndTime: 2016-12-04 23:24:55 +0900 JST,
  MetricName: "CPUUtilization",
  Namespace: "AWS/EC2",
  Period: 60,
  StartTime: 2016-12-04 23:19:55 +0900 JST,
  Statistics: ["Average"]
}
```

Run command below to view full option list.

```
$ metrin check --help
```

#### `print`

Print metric statistic value with given format (`--template`).

Example: Metric path, value, timestamp, tab separated (Default template)

```
$ metrin --namespace AWS/EC2 --metric-name CPUUtilization --statistic Average --start-time -900 --period 300 --dimension InstanceId:i-abcd1234abcd12345 print
CloudWatch.InstanceId.i-abcd1234abcd12345.CPUUtilization.Average	1.566	1480861800
CloudWatch.InstanceId.i-abcd1234abcd12345.CPUUtilization.Average	1.536	1480862100
```

Example: Single value only output (for Zabbix, etc.)

```
$ metrin --namespace AWS/EC2 --metric-name CPUUtilization --statistic Average --start-time -900 --period 300 --dimension InstanceId:i-abcd1234abcd12345 print --template '{{.Datapoint.Average}}' --last-value-only
1.536
```

Run command below to view full option list.

```
$ metrin print --help
```

This feature is using [text/template](https://golang.org/pkg/text/template/) package.

Additional functions are `unixtime` (Convert `.Datapoint.Timestamp` to UNIX timestamp), `deref` (Convert `*float64` to `float64`, combine usage with `printf`), `getvalue` (Get statistic value dynamically).

See [print.go](lib/print.go).

#### `debug`

For troubleshooting.

## FAQ.

### `No datapoints`?

Adjust `--start-time`, `--end-time`, `--period`.

### Multiple dimensions?

```
--dimension dim1key:dim1value --dimension dim2key:dim2value
```
