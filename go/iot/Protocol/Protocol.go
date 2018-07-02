// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package Protocol

import "go.opencensus.io/stats/view"

/*
Typical example for Registration request would be as below:
{
	"ProjectId" : "project-id"
	"View" : {
		"Name" : "my.org/views/protocol_test"
		"Description" : "View for Protocol Test"
	}
	"Aggregation" : {
		"AggregationType" : "LastValue" / "Sum" / "Count"
		"AggregationValue" : []
	}
	"Measure":{
		"Name" : "my.org/measure/Measure_Test"
		"Descrption" : "Measure Test"
		"Unit" : "1" / "By" / "ms"
		"MeasureType": "int64" / "float64"
	}
	"ReportPeriod" : 1

}

Note that the distribution is not supported yet.


Typical example for sending data request would be as below:
{
	"Measure":{
		"Name": "my.org/measure/Measure_Test"
		"MeasureType": "int64" / "float64"
		"MeasureValue": "1"
	}
}

Typical example for the response from Raspberry Pi would be as below:
{
	"Code": 200 (OK) / 404 (Fail)
	"Info": "Registration Successfully!"
}

*/

// TODO: In JSON, we could use this array. Without JSON, we need to implement the parse
const (
	REGISTRATION = 0
	RECORD       = 1
	OK           = 200
	FAIL         = 404
)

type Argument struct {
	ArgumentType int
	ProjectId    string
	DeviceId string
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
	AggregationType  string
	AggregationValue []float64
}

type Response struct {
	Code int
	Info string
}
