package test

import (
	"testing"

	parser2 "github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/parser"
)

var jsonParser parser2.JsonParser

// Test Example without any whitespace
// {"Name":"opencensus.io/measure/Temperature","Value":"23.72","Tags":{"ArduinoId":"Arduino-1","Date":"2018-07-02"}}
func Test_jsonparser_normal(t *testing.T) {
	var testEg string = "{\"Name\":\"opencensus.io/measure/Temperature\",\"Value\":\"23.72\",\"Tags\":{\"ArduinoId\":\"Arduino-1\",\"Date\":\"2018-07-02\"}}"
	result, err := jsonParser.Decode([]byte(testEg))
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
func Test_jsonparser_normal_with_space(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Value\":\"23.72\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } }"
	result, err := jsonParser.Decode([]byte(testEg))
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
func Test_jsonparser_normal_unordered(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Value\":\"23.72\" }"
	result, err := jsonParser.Decode([]byte(testEg))
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
func Test_jsonparser_abnormal_unpaired_brackets(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Value\":\"23.72\" }}"
	_, err := jsonParser.Decode([]byte(testEg))
	if err == nil {
		t.Error("There is one more bracket in the end of input string")
	}
}

// Test Example with different keys as defined in the protocol
// { "Name": "opencensus.io/measure/Temperature" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Meaurement":"23.72" }}
func Test_jsonparser_abnormal_wrong_key(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Valu\":\"23.72\" }"
	_, err := jsonParser.Decode([]byte(testEg))
	if err != nil {
		// Compared to the text parser, JsonParser won't throw out any error. Instead, it would leave the Value alone.
		t.Error("It should be Value instead of Valu")
	}
}

// Test Example with different keys as defined in the protocol
// { "Name": "opencensus.io/measure/Temperature" , "Tags":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Meaurement":"23.72" ,}}
func Test_jsonparser_abnormal_more_comma(t *testing.T) {
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tags\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Value\":\"23.72\" ,}"
	_, err := jsonParser.Decode([]byte(testEg))
	if err == nil {
		// Compared to the text parser, JsonParser won't tolerate any meaningless character in the end or beginning
		t.Error("It should thow out an error since there is a comma in the end")
	}
}
