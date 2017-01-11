package metrin

import (
	"bytes"
	"html/template"
	"reflect"

	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// BuildPrintStringInput - includes params and datapoints
type BuildPrintStringInput struct {
	Params         *cloudwatch.GetMetricStatisticsInput
	Datapoints     []*cloudwatch.Datapoint
	TemplateString string
}

// TemplateInput - input type for each template execution
type TemplateInput struct {
	Params    *cloudwatch.GetMetricStatisticsInput
	Datapoint *cloudwatch.Datapoint
}

// BuildPrintStrings - returns slice of built string
func BuildPrintStrings(input BuildPrintStringInput) []string {
	var strings []string

	buildTemplate := template.New("")

	buildTemplate.Funcs(template.FuncMap{
		"unixtime": func(t time.Time) int64 { return t.Unix() },
		"deref":    func(v *float64) float64 { return *v },

		"getvalue": func(datapoint *cloudwatch.Datapoint, params *cloudwatch.GetMetricStatisticsInput, statIndex int) *float64 {
			r := reflect.Indirect(reflect.ValueOf(datapoint))
			f := r.FieldByName(*params.Statistics[statIndex]).Interface().(*float64)

			return f
		},
	})

	template.Must(buildTemplate.Parse(input.TemplateString))

	for i := range input.Datapoints {
		datapoint := input.Datapoints[i]
		buffer := new(bytes.Buffer)

		buildTemplate.Execute(buffer, TemplateInput{
			Params:    input.Params,
			Datapoint: datapoint,
		})

		strings = append(strings, buffer.String())
	}

	return strings
}
