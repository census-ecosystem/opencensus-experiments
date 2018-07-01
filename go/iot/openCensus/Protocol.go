package openCensus

import "go.opencensus.io/stats/view"

/*
Typical Argument for Registration would be as below:
{
	"projectId" : "opencensus-java-stats-demo-app"
	"view" : {
		"Name" : "my.org/views/protocol_test"
		"Description" : "View for Protocol Test"
	}
	"aggregation" : {
		"aggregationType" : "LastValue"
		"aggregationValue" : []
	}
	"measure":{
		"name" : "my.org/measure/Measure_Test"
		"descrption" : "Measure Test"
		"unit" : "1"
		"measureType": "int64"
		"measureValue": ""
	}
	"reportPeriod" : 1

}


Typical Argument for Data would be as below:
{
	"projectId" :
	"view" :
	"aggregation" :
	"measure":{
		"name" : "my.org/measure/Measure_Test"
		"descrption":
		"unit" :
		"measureType": "int64"
		"measureValue": "1"
	}
	"reportPeriod" :
}

 */
type Argument struct {
	projectId    string
	view         view.View
	Aggregation  AggregationArgument
	Measure      MeasureArgument
	reportPeriod int
}

type MeasureArgument struct {
	Name string
	Description string
	Unit string

	// We judge the measure Type based on the users. Which means we assume that user would not input malformed data
	MeasureType string
	MeasureValue string
}

type AggregationArgument struct {
	AggregationType string
	// TODO: In JSON, we could use this array. Without JSON, we need to implement the parse
	AggregationValue []float64
}

