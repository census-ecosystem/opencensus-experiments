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

	"github.com/census-ecosystem/opencensus-experiments/go/iot/openCensus"
	"github.com/huin/goserial"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	testMeasure  = stats.Int64("my.org/measure/Measure_Test", "Measure Test", stats.UnitDimensionless)
	reportPeriod = 1
	ardiunoKey   tag.Key
	dateKey      tag.Key
)

func findArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for what is mostly likely the Arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbserial") ||
			strings.Contains(f.Name(), "ttyACM") {
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}

func main() {
	c := &goserial.Config{Name: findArduino(), Baud: 9600}
	var slave openCensus.Slave
	var census openCensus.OpenCensusBase
	projectId := os.Getenv("PROJECTID")
	if projectId == "" {
		log.Fatal("Cannot detect PROJECTID in the system environment.\n")
	} else {
		log.Printf("Project Id is set to be %s\n", projectId)
	}
	census.Initialize(projectId, reportPeriod)
	census.ViewRegistration(&view.View{
		Name:        "my.org/views/protocol_test",
		Description: "View for Protocol Test",
		Aggregation: view.LastValue(),
		Measure:     testMeasure,
		TagKeys:     getExampleKey(),
	})
	slave.Initialize(c)
	slave.Subscribe(census)
	slave.Collect(2 * time.Second)
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
