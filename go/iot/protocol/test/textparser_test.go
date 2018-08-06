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

package test

import (
	parser2 "github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/parser"
	"testing"
)

var textParser parser2.TextParser

// Test Example without any whitespace
// {"Name":"opencensus.io/measure/Temperature","Value":"23.72","Tags":{"ArduinoId":"Arduino-1","Date":"2018-07-02"}}
func Test_textparser_normal(t *testing.T) {
	var testEg string = "{\"Name\":\"opencensus.io/measure/Temperature\",\"Value\":\"23.72\",\"Tags\":{\"ArduinoId\":\"Arduino-1\",\"Date\":\"2018-07-02\"}}"
	result, err := textParser.DecodeMeasurement([]byte(testEg))
	if err != nil {
		t.Error(err)
	} else {
		if result.Name != "opencensus.io/measure/Temperature" {
			t.Error("Name of parse result is wrong as ", result.Name, " Correct answer is opencensus.io/measure/Temperature")
		}

		if result.Value != "23.72" {
			t.Error("Value of parse result is wrong as ", result.Value, " Correct answer is 23.72")
		}

		if result.Tags["ArduinoId"] != "Arduino-1" {
			t.Error("Tags ArduinoId of parse result is wrong as ", result.Tags["ArduinoId"], " Correct answer is Arduino-1")
		}

		if result.Tags["Date"] != "2018-07-02" {
			t.Error("Tags ArduinoId of parse result is wrong as ", result.Tags["Date"], " Correct answer is 2018-07-02")
		}
	}
}

// Test Example with some whitespaces
// { "Name": "opencensus.io/measure/Temperature" , "Value":"23.72" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } }
func Test_textparser_normal_with_space(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Value\":\"23.72\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } }"
	result, err := textParser.DecodeMeasurement([]byte(testEg))
	if err != nil {
		t.Error(err)
	} else {
		if result.Name != "opencensus.io/measure/Temperature" {
			t.Error("Name of parse result is wrong as ", result.Name, " Correct answer is opencensus.io/measure/Temperature")
		}

		if result.Value != "23.72" {
			t.Error("Value of parse result is wrong as ", result.Value, " Correct answer is 23.72")
		}

		if result.Tags["ArduinoId"] != "Arduino-1" {
			t.Error("Tags ArduinoId of parse result is wrong as ", result.Tags["ArduinoId"], " Correct answer is Arduino-1")
		}

		if result.Tags["Date"] != "2018-07-02" {
			t.Error("Tags ArduinoId of parse result is wrong as ", result.Tags["Date"], " Correct answer is 2018-07-02")
		}
	}
}

// Test Example with different order of key-value pairs
// { "Name": "opencensus.io/measure/Temperature" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Value":"23.72" }
func Test_textparser_normal_unordered(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Value\":\"23.72\" }"
	result, err := textParser.DecodeMeasurement([]byte(testEg))
	if err != nil {
		t.Error(err)
	} else {
		if result.Name != "opencensus.io/measure/Temperature" {
			t.Error("Name of parse result is wrong as ", result.Name, " Correct answer is opencensus.io/measure/Temperature")
		}

		if result.Value != "23.72" {
			t.Error("Value of parse result is wrong as ", result.Value, " Correct answer is 23.72")
		}

		if result.Tags["ArduinoId"] != "Arduino-1" {
			t.Error("Tags ArduinoId of parse result is wrong as ", result.Tags["ArduinoId"], " Correct answer is Arduino-1")
		}

		if result.Tags["Date"] != "2018-07-02" {
			t.Error("Tags ArduinoId of parse result is wrong as ", result.Tags["Date"], " Correct answer is 2018-07-02")
		}
	}
}

// Test Example with unpaired brackets in the end of input string
// { "Name": "opencensus.io/measure/Temperature" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Value":"23.72" }}
func Test_textparser_abnormal_unpaired_brackets(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Value\":\"23.72\" }}"
	_, err := textParser.DecodeMeasurement([]byte(testEg))
	if err == nil {
		t.Error("There is one more bracket in the end of input string")
	}
}

// Test Example with different keys as defined in the protocol
// { "Name": "opencensus.io/measure/Temperature" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Meaurement":"23.72" }}
func Test_textparser_abnormal_wrong_key(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Valu\":\"23.72\", }"
	_, err := textParser.DecodeMeasurement([]byte(testEg))
	if err == nil {
		t.Error("It should be Value instead of Valu")
	}
}

// Test Example with different keys as defined in the protocol
// {, "Name": "opencensus.io/measure/Temperature" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Meaurement":"23.72" ,}}
func Test_textparser_abnormal_more_comma(t *testing.T) {
	var testEg string = "{, \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Value\":\"23.72\" ,}"
	_, err := textParser.DecodeMeasurement([]byte(testEg))
	if err != nil {
		// Compared to the JsonParser, textParser will tolerate meaningless character in the end or beginning
		t.Error("It shouldn't thow out an error even if there is a comma in the end")
	}
}

