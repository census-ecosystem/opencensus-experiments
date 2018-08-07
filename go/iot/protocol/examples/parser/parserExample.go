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

package main

import (
	"fmt"
	parser2 "github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/parser"
	)

func main() {
	var example string = "{\"Name\":\"opencensus.io/measure/Temperature\",\"Value\":\"23.72\",\"Tags\":{\"ArduinoId\":\"Arduino-1\",\"Date\":\"2018-07-02\"}}"
	var parser parser2.TextParser
	result, err := parser.DecodeMeasurement([]byte(example))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Name: ", result.Name)
		fmt.Println("Measurement Value: ", result.Value)
		for k, v := range result.Tags {
			fmt.Println("Key: ", k, " Value: ", v)
		}
	}
}
