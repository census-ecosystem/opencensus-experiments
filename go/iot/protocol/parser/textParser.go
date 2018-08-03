package parser

import (
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
	"github.com/pkg/errors"
	"strings"
)

type TextParser struct {
}

func (parser *TextParser) Parse(input []byte) (protocol.MeasureArgument, error) {
	var output protocol.MeasureArgument
	// parseResult is only the transition result of parse. It is in the form of map[string]interface{}
	// Using the interface{} can make it more general since we might change the protocol in the future.
	parseResult, err := parser.helper(string(input))
	if (err != nil){
		return output, err
	}
	var ok bool
	output.Name, ok= parseResult["Name"].(string)
	if ok == false{
		return output, errors.Errorf("Value for Name is not valid string")
	}

	output.Value, ok = parseResult["Value"].(string)
	if ok == false{
		return output, errors.Errorf("Value for Value is not valid string")
	}

	output.Tags, ok = parseResult["Value"].(map[string]string)
	if ok == false{
		return output, errors.Errorf("Value for Tag key pair is not valid string")
	}

	return output, nil
}

// Typical Input:
// {"Name":"opencensus.io/measure/Temperature","Measurement":"23.72","Tag":{"ArduinoId":"Arduino-1","Date":"2018-07-02"}}
func (parser *TextParser) helper(ss string) (map[string]interface{}, error){

	var res map[string]interface{}

	// First of all, trim all the leading and trailing whitespaces of the input string
	ss = strings.Trim(ss, " ")

	// Second of all, the trimmed string should has a leading and trailing bracket, which forms a valid bracket
	if ss[0] != '{' || ss[len(ss) - 1] != '}'{
		return res, errors.Errorf("Data not started or ended with { / }")
	}

	// Then trim the leading and trailing bracket of the string
	ss = ss[1 : len(ss) - 1]

	// Key-value pair in the string are separated by comma.
	sspairs := strings.Split(ss, ",")

	// If there are not key-value pair in the string, directly return a empty map
	for _, sspair := range sspairs{
		// Key and value in the pair are separated by colons.
		kvpair := strings.Split(sspair, ":")

		if (len(kvpair) != 2){
			return res, errors.Errorf("More than one colon in the key/value pairs")
		}

		// before handling the key and value, we need trim all the leading and trailing whitespaces of them
		key := strings.Trim(kvpair[0], " ")
		value := strings.Trim(kvpair[1], " ")

		// Since every key / value has a leading / trailing bracket, its size has to be larger than 2.
		if len(key) <= 2 || len(value) <= 2 {
			return res, errors.Errorf("Key or Value is Empty")
		}

		// Every key / value has a leading / trailing bracket, otherwise it's invalid.
		if key[0] != '"' || key[len(key) - 1] != '"' {
			return res, errors.Errorf("Key not started or ended with quataion marks")
		}

		// Character on the both side of value has to form a pair of quataion marks or brackets
		if !((value[0] == '"' && value[len(value) - 1] == '"') || (value[0] == '{' && value[len(value) - 1] == '}')){
			return res, errors.Errorf("Value not started / ended with quatation marks / brackets")
		}

		if (value[0] == '"' && value[len(value) - 1] == '"'){
			// Value is in the form of string
			key = key[1 : len(key) - 2]
			value = value[1 : len(value) - 2]
			res[key] = value
		} else{
			// value is another map of key-value pairs
			subvalue, err := parser.helper(value)
			if err != nil {
				return res, err
			} else{
				res[key] = subvalue
			}
		}
	}

	return res, nil
}