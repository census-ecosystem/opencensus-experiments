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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	"fmt"

	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/opencensus"
	"github.com/huin/goserial"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)


var (
	testMeasure  = stats.Float64("opencensus.io/measure/Temperature", "Measure Test", stats.UnitDimensionless)
	reportPeriod = 1
	ardiunoKey   tag.Key
	dateKey      tag.Key
)

func findArduino() []string {
	var arduinoList []string
	contents, _ := ioutil.ReadDir("/dev")
	// Look for what is mostly likely the Arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbserial") ||
			strings.Contains(f.Name(), "ttyACM") {
			arduinoList = append(arduinoList, "/dev/" + f.Name())
		}
	}

	return arduinoList
}

func main() {
	projectId := os.Getenv("PROJECTID")
	if projectId == "" {
		log.Fatal("Cannot detect PROJECTID in the system environment.\n")
	} else {
		log.Printf("Project Id is set to be %s\n", projectId)
	}
	var census opencensus.OpenCensusBase
	census.Initialize(projectId, reportPeriod)
	census.ViewRegistration(&view.View{
		Name:        "opencensus.io/views/protocol_test",
		Description: "View for Protocol Test",
		Aggregation: view.LastValue(),
		Measure:     testMeasure,
		TagKeys:     getExampleKey(),
	})

	for _, slaveName := range findArduino(){
		c := &goserial.Config{Name: slaveName, Baud: 9600}
		var slave opencensus.Slave
		slave.Initialize(c)
		slave.Subscribe(census)
		slave.Collect(2 * time.Second)
	}
	for true{
		fmt.Print()
	}

}

func getExampleKey() []tag.Key {
	var exampleKey []tag.Key
	if ardiunoKey, err := tag.NewKey("ArduinoId"); err == nil {
		exampleKey = append(exampleKey, ardiunoKey)
	} else {
		log.Fatal("Unable to create new Tag Key\n")
	}
	if dateKey, err := tag.NewKey("Date"); err == nil {
		exampleKey = append(exampleKey, dateKey)
	} else {
		log.Fatal("Unable to create new Tag Key\n")
	}
	return exampleKey
}
