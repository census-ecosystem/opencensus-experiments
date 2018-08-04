package test

import (
	"testing"
	parser2 "github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/parser"
	)
var parser parser2.TextParser

// Test Example without any whitespace
// {"Name":"opencensus.io/measure/Temperature","Measurement":"23.72","Tag":{"ArduinoId":"Arduino-1","Date":"2018-07-02"}}
func Test_parser_normal (t *testing.T){
	var testEg string = "{\"Name\":\"opencensus.io/measure/Temperature\",\"Measurement\":\"23.72\",\"Tag\":{\"ArduinoId\":\"Arduino-1\",\"Date\":\"2018-07-02\"}}"
	result, err := parser.Parse([]byte(testEg))
	if err != nil {
		t.Error(err)
	} else{
		if result.Name != "opencensus.io/measure/Temperature"{
			t.Error("Name of parse result is wrong as ", result.Name, " Correct answer is opencensus.io/measure/Temperature")
		}

		if result.Value != "23.72"{
			t.Error("Value of parse result is wrong as ", result.Value, " Correct answer is 23.72")
		}

		if result.Tags["ArduinoId"] != "Arduino-1"{
			t.Error("Tag ArduinoId of parse result is wrong as ", result.Tags["ArduinoId"], " Correct answer is Arduino-1")
		}

		if result.Tags["Date"] != "2018-07-02"{
			t.Error("Tag ArduinoId of parse result is wrong as ", result.Tags["Date"], " Correct answer is 2018-07-02")
		}
	}
}

// Test Example with some whitespaces
// { "Name": "opencensus.io/measure/Temperature" , "Measurement":"23.72" , "Tag":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } }
func Test_parser_normal_with_space (t *testing.T){
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Measurement\":\"23.72\" , \"Tag\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } }"
	result, err := parser.Parse([]byte(testEg))
	if err != nil {
		t.Error(err)
	} else{
		if result.Name != "opencensus.io/measure/Temperature"{
			t.Error("Name of parse result is wrong as ", result.Name, " Correct answer is opencensus.io/measure/Temperature")
		}

		if result.Value != "23.72"{
			t.Error("Value of parse result is wrong as ", result.Value, " Correct answer is 23.72")
		}

		if result.Tags["ArduinoId"] != "Arduino-1"{
			t.Error("Tag ArduinoId of parse result is wrong as ", result.Tags["ArduinoId"], " Correct answer is Arduino-1")
		}

		if result.Tags["Date"] != "2018-07-02"{
			t.Error("Tag ArduinoId of parse result is wrong as ", result.Tags["Date"], " Correct answer is 2018-07-02")
		}
	}
}

// Test Example with different order of key-value pairs
// { "Name": "opencensus.io/measure/Temperature" , "Tag":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Measurement":"23.72" }
func Test_parser_normal_unordered (t *testing.T){
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tag\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Measurement\":\"23.72\" }"
	result, err := parser.Parse([]byte(testEg))
	if err != nil {
		t.Error(err)
	} else{
		if result.Name != "opencensus.io/measure/Temperature"{
			t.Error("Name of parse result is wrong as ", result.Name, " Correct answer is opencensus.io/measure/Temperature")
		}

		if result.Value != "23.72"{
			t.Error("Value of parse result is wrong as ", result.Value, " Correct answer is 23.72")
		}

		if result.Tags["ArduinoId"] != "Arduino-1"{
			t.Error("Tag ArduinoId of parse result is wrong as ", result.Tags["ArduinoId"], " Correct answer is Arduino-1")
		}

		if result.Tags["Date"] != "2018-07-02"{
			t.Error("Tag ArduinoId of parse result is wrong as ", result.Tags["Date"], " Correct answer is 2018-07-02")
		}
	}
}

// Test Example with unpaired brackets in the end of input string
// { "Name": "opencensus.io/measure/Temperature" , "Tag":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Measurement":"23.72" }}
func Test_parser_abnormal_unpaired_brackets (t *testing.T){
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tag\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Measurement\":\"23.72\" }}"
	_, err := parser.Parse([]byte(testEg))
	if err != nil {
	} else{
		t.Error("There is one more bracket in the end of input string")
	}
}

// Test Example with different keys as defined in the protocol
// { "Name": "opencensus.io/measure/Temperature" , "Tag":{"ArduinoId" : "Arduino-1" , "Date" : "2018-07-02"  } , "Meaurement":"23.72" }}
func Test_parser_abnormal_wrong_key (t *testing.T){
	var testEg string = "{ \"Name\": \"opencensus.io/measure/Temperature\" , \"Tag\" : { \"ArduinoId\" : \"Arduino-1\" , \"Date\" : \"2018-07-02\"  } , \"Meaurement\":\"23.72\" }"
	_, err := parser.Parse([]byte(testEg))
	if err != nil {
	} else{
		t.Error("It should be Measurement instead of Meaurement")
	}
}