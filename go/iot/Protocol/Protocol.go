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
const (
	REGISTRATION = 0
	RECORD       = 1
	OK           = 200
	FAIL         = 404
)

type MeasureArgument struct {
	Name string

	// We judge the measure Type based on the users.
	// Which means we assume that user would not input malformed data
	MeasureType  string
	MeasureValue string
	TagValues    []string
}

type Response struct {
	Code int
	Info string
}
