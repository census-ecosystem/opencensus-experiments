// Copyright 2017, OpenCensus Authors
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

//  Program iot uploads sensor data to monitoring backends by using the OpenCensus framework.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	high = 1
)

var (
	gpioVoltage = stats.Int64("my.org/measure/gpio_voltage_level", "level of voltage", stats.UnitDimensionless)
)

func main() {
	ctx := context.Background()
	// TODO: It takes around one minute to detect the full edge of voltage change. Needs to tune the report period
	projectId := os.Getenv("PROJECTID")
	if projectId == "" {
		log.Fatal("Cannot detect PROJECTID in the system environment.\n")
	} else {
		log.Printf("Project Id is set to be %s\n", projectId)
	}
	initOpenCensus(projectId, 1)
	r := raspi.NewAdaptor()
	myGPIO := gpio.NewDirectPinDriver(r, "11")
	led := gpio.NewLedDriver(r, "7")
	work := func() {
		gobot.Every(1*time.Second, func() {
			voltage, err := myGPIO.DigitalRead()
			if err != nil {
				log.Fatalf("Error with Reading Voltage on the Raspberry Pi Pin 11")
			} else {
				if voltage == high {
					led.On()
				} else {
					led.Off()
				}
			}
			recordVoltage(ctx, int64(voltage))
		})
	}

	robot := gobot.NewRobot("PinVoltageCollection",
		[]gobot.Connection{r},
		[]gobot.Device{myGPIO, led},
		work,
	)
	robot.Start()
}

func initOpenCensus(projectId string, reportPeriod int) {
	// Collected view data will be reported to Stackdriver Monitoring API
	// via the Stackdriver exporter.
	//
	// In order to use the Stackdriver exporter, enable Stackdriver Monitoring API
	// at https://console.cloud.google.com/apis/dashboard.
	//
	// Once API is enabled, you can use Google Application Default Credentials
	// to setup the authorization.
	// See https://developers.google.com/identity/protocols/application-default-credentials
	// for more details.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectId, // Google Cloud Console project ID.
	})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(exporter)

	// Set reporting period to report data at every second.
	view.SetReportingPeriod(time.Second * time.Duration(reportPeriod))

	// Create view to see the processed video size cumulatively.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Register(&view.View{
		Name:        "my.org/views/gpio_voltage_instant",
		Description: "voltage level on GPIO over time",
		Measure:     gpioVoltage,
		Aggregation: view.LastValue(),
	}); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}
}

func recordVoltage(ctx context.Context, voltage int64) {
	stats.Record(ctx, gpioVoltage.M(voltage))
}
