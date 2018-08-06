package parser

import (
	"strings"

	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
	"github.com/pkg/errors"
)

type TextParser struct {
}

// In this function, it firstly transform the input byte stream into a map of string -> interface{}.
// If it manages so, it would extract the required variable as defined in the protocol.
// Otherwise it would return an un-empty error.
func (parser *TextParser) Decode(input []byte) (protocol.MeasureArgument, error) {
	var output protocol.MeasureArgument
	// parseResult is only the transition result of parse. It is in the form of map[string]interface{}
	// Using the interface{} can make it more general since we might change the protocol in the future.
	var parseResult map[string]interface{} = make(map[string]interface{})

	ss := string(input)
	// First of all, find the outermost layer of bracket.
	firstBracketPos := strings.Index(ss, "{")
	lastBracketPos := strings.LastIndex(ss, "}")
	if firstBracketPos == -1 || lastBracketPos == -1 {
		return output, errors.Errorf("Unpaired brackets in the two end of input string")
	}
	// Then trim the leading and trailing bracket of the string
	ss = ss[firstBracketPos+1 : lastBracketPos]
	err := parser.helper(ss, parseResult)

	// If the error is not empty, return it.
	if err != nil {
		return output, err
	}
	var ok bool
	output.Name, ok = parseResult["Name"].(string)
	if ok == false {
		return output, errors.Errorf("Value for Name is not valid string")
	}

	output.Value, ok = parseResult["Value"].(string)
	if ok == false {
		return output, errors.Errorf("Value for Measurement is not valid string")
	}

	output.Tags = make(map[string]string)

	// Since we would not assert the type of map[string]interface{} to map[string]string, we need to copy one by one
	for k, v := range parseResult["Tags"].(map[string]interface{}) {
		output.Tags[k], ok = v.(string)
		if ok == false {
			//fmt.Print(test)
			return output, errors.Errorf("Value for Tags key pair is not valid string")
		}
	}

	return output, nil
}

// The input string of helper function is already trimmed of the leading and trailing bracket.
func (parser *TextParser) helper(ss string, res map[string]interface{}) error {

	// Judge whether there are nested brackets
	var pos int = strings.Index(ss, "{")

	if pos == -1 {
		// No nested brackets. We could parse the string through the helper function.
		err := parser.parseWithNoBracket(ss, res)
		return err
	} else {
		// Nested Bracket exists in the string. Currently we know the start of nested string, we firstly find its end
		count := 1
		lastBracket := pos + 1
		for ; lastBracket < len(ss); lastBracket = lastBracket + 1 {
			if ss[lastBracket] == '}' {
				count = count - 1
			}
			if ss[lastBracket] == '{' {
				count = count + 1
			}

			if count == 0 {
				break
			}
		}

		// If the count is not zero after we traverse the whole string, it means that left and right brackets are not paired
		if count != 0 {
			return errors.Errorf("Nested bracket unvalid")
		}

		var nestedResult map[string]interface{} = make(map[string]interface{})

		// Recursively handle the string without leading and trailing brackets
		err := parser.helper(ss[pos+1:lastBracket], nestedResult)
		if err != nil {
			return nil
		} else {
			// We would need to find the key of this nestedResult
			start := pos - 1
			for start >= 0 && ss[start] != '"' {
				start = start - 1
			}

			if start == -1 {
				return errors.Errorf("No key found!")
			}

			end := start - 1

			for end >= 0 && ss[end] != '"' {
				end = end - 1
			}

			if end == -1 {
				return errors.Errorf("No pair of quatation mark found")
			}

			key := ss[end+1 : start]

			if len(key) == 0 {
				return errors.Errorf("Key is empty string")
			}

			res[key] = nestedResult

			// Before 'end', there is no bracket, which means there is no nested brackets. So we call the helper function
			if err := parser.parseWithNoBracket(ss[0:end], res); err != nil {
				return err
			}

			if lastBracket+1 == len(ss) {
				// The end of string
				return nil
			} else {
				// There are key-value pairs left behind. We should continue to parse the rest of string
				if err := parser.helper(ss[lastBracket+1:len(ss)], res); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// This function parses the string without any nested map as the value
// The input map is the reference to the parse result
func (parser *TextParser) parseWithNoBracket(ss string, res map[string]interface{}) error {

	ss = strings.Trim(ss, " ")
	ss = strings.Trim(ss, ",")
	ss = strings.Trim(ss, " ")

	if len(ss) == 0 {
		// Remaining characters are all whitespaces
		return nil
	}
	// Key-value pair in the string are separated by comma.
	sspairs := strings.Split(ss, ",")

	// If there are not key-value pair in the string, directly return a empty map
	for _, sspair := range sspairs {
		// Key and value in the pair are separated by colons.
		kvpair := strings.Split(sspair, ":")

		if len(kvpair) != 2 {
			return errors.Errorf("More than one colon in the key/value pairs")
		}

		// before handling the key and value, we need trim all the leading and trailing whitespaces of them
		key := strings.Trim(kvpair[0], " ")
		value := strings.Trim(kvpair[1], " ")

		// Since every key / value has a leading / trailing bracket, its size has to be larger than 2.
		if len(key) <= 2 || len(value) <= 2 {
			return errors.Errorf("Key or Value is Empty")
		}

		// Every key / value has a leading / trailing bracket, otherwise it's invalid.
		if key[0] != '"' || key[len(key)-1] != '"' {
			return errors.Errorf("Key not started or ended with quataion marks")
		}

		// Character on the both side of value has to form a pair of quataion marks
		if value[0] != '"' || value[len(value)-1] != '"' {
			return errors.Errorf("Value not started / ended with quatation marks / brackets")
		}

		key = key[1 : len(key)-1]
		value = value[1 : len(value)-1]
		res[key] = value
	}

	return nil
}


func (parser *TextParser) Encode(myResponse *protocol.Response) ([]byte, error) {
	var res string = "\"Code\":" + string(myResponse.Code) + "\"Info\":" + myResponse.Info
	return []byte(res), nil
}