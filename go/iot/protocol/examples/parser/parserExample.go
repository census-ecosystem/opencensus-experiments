package main

import (
	"fmt"
	parser2 "github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/parser"
)

func main() {
	var example string = "{\"Name\":\"opencensus.io/measure/Temperature\",\"Measurement\":\"23.72\",\"Tag\":{\"ArduinoId\":\"Arduino-1\",\"Date\":\"2018-07-02\"}}"
	var parser parser2.TextParser
	result, err := parser.Parse([]byte(example))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Name: ", result.Name)
		fmt.Println("Measurement Value: ", result.Measurement)
		for k, v := range result.Tag {
			fmt.Println("Key: ", k, " Value: ", v)
		}
	}
}
