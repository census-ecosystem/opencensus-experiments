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

// Program iot uploads data for same metric but different sensors
// to monitoring backend by using the OpenCensus framework.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/d2r2/go-dht"
	"github.com/d2r2/go-logger"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

// Create measures. We concurrently upload the data that is from different sensor under the main.go program.
// In this way, we could easily analyze the difference between sensors on the same metrics.
var (
	temperatureMeasureCompare = stats.Float64("my.org/measure/temperature_svl_mp1_7c3c", "temperature", stats.UnitDimensionless)

	lgCompare = logger.NewPackageLogger("main",
		logger.DebugLevel,
		// logger.InfoLevel,
	)
)

// The DHT11 sensor in this program is supposed to connected to the GPIO17 on the raspberry Pi.
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

	go RecordTemperatureHumidityCompare(ctx, 17)
	board := raspi.NewAdaptor()
	ads1015 := i2c.NewADS1015Driver(board)

	work := func() {
	}

	robot := gobot.NewRobot("sensorDataCollection",
		[]gobot.Connection{board},
		[]gobot.Device{ads1015},
		work,
	)
	robot.Start()
}

// For every five seconds, record the temperature and humidity sensor data.
// Print the collected data on the console.
func RecordTemperatureHumidityCompare(ctx context.Context, pin int) {
	for range time.Tick(5 * time.Second) {
		defer logger.FinalizeLogger()
		// Uncomment/comment next line to suppress/increase verbosity of output
		logger.ChangePackageLogLevel("dht", logger.InfoLevel)

		sensorType := dht.DHT11
		// Read DHT11 sensor data from pin 4, retrying 50 times in case of failure.
		// You may enable "boost GPIO performance" parameter, if your device is old
		// as Raspberry PI 1 (this will require root privileges). You can switch off
		// "boost GPIO performance" parameter for old devices, but it may increase
		// retry attempts. Play with this parameter.
		temperature, humidity, retried, err :=
			dht.ReadDHTxxWithRetry(sensorType, pin, false, 50)
		if err != nil {
			lgCompare.Fatal(err)
		}
		// print temperature and humidity
		lgCompare.Infof("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)",
			sensorType, temperature, humidity, retried)
		stats.Record(ctx, temperatureMeasureCompare.M(float64(temperature)))
	}
}

// Initialize the openCensus framework.
// If there is anything wrong with the registration, directly throw a fatal error.
// Upload the sensor data to the same view on the backend driver.
func initOpenCensusCompare(projectId string, reportPeriod int) {
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

	// Create view to see the temperature instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Register(&view.View{
		Name:        "my.org/views/temperature_instant",
		Description: "temperature over time",
		Measure:     temperatureMeasureCompare,
		Aggregation: view.LastValue(),
	}); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}
}
