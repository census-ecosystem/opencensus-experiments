package Protocol

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

// TODO: In JSON, we could use this array. Without JSON, we need to implement the parse
const (
	REGISTRATION = 0
	RECORD       = 1
)

type Argument struct {
	ArgumentType int
	ProjectId    string
	View         view.View
	Aggregation  AggregationArgument
	Measure      MeasureArgument
	ReportPeriod int
}

type MeasureArgument struct {
	Name        string
	Description string
	Unit        string

	// We judge the measure Type based on the users. Which means we assume that user would not input malformed data
	MeasureType  string
	MeasureValue string
}

type AggregationArgument struct {
	AggregationType string
	AggregationValue []float64
}

type Response struct {
	Code int
	Info string
}
