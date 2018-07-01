package driver

import (
	"go.opencensus.io/stats/view"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/openCensus"
	"go.opencensus.io/stats"
	v "go.opencensus.io/stats/view"
	"log"
)

func ViewParse(view *view.View, argument *openCensus.Argument) error{
	// Name and Description of the view should be already updated by the JSON.
	// TODO: If not using JSON, we should update the name and description by ourseleves
	measureString := argument.Measure
	switch measureString.MeasureType {
	case "int64":
		view.Measure = stats.Int64(measureString.Name, measureString.Description, measureString.Unit)
	case "float64":
		view.Measure = stats.Float64(measureString.Name, measureString.Description, measureString.Unit)
	default:
		log.Fatal("Not Supported Measure Type\n")
	}

	aggregationString := argument.Aggregation
	switch aggregationString.AggregationType {
	case "LastValue":
		view.Aggregation = v.LastValue()
	case "Sum":
		view.Aggregation = v.Sum()
	case "Count":
		view.Aggregation = v.Count()
	case "Distribution":
		// TODO: Don't know how to transfer from Slice to ... argument
		view.Aggregation = v.Distribution(aggregationString.AggregationValue[0])
		//view.Aggregation = v.Distribution(aggregationString.AggregationValue[0], aggregationString.AggregationValue[1:]...)
	default:
		log.Fatal("Not Supported Aggregation Type\n")
	}
	return nil
}
