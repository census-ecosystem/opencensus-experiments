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

// Starter codes for the iot application under the OpenCensus framework

package main

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/d2r2/go-dht"
	"github.com/d2r2/go-logger"
	"go.opencensus.io/stats/view"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"time"
	"go.opencensus.io/stats"
)

// Create measures. The program will record measures for the voltage
// level on the specific GPIO
var (
	humidityMeasureCompare          = stats.Float64("my.org/measure/humidity_svl_mp1_7c3c", "humidity", stats.UnitDimensionless)
	temperatureMeasureCompare       = stats.Float64("my.org/measure/temperature_svl_mp1_7c3c", "temperature", stats.UnitDimensionless)

	lgCompare = logger.NewPackageLogger("main",
		logger.DebugLevel,
		// logger.InfoLevel,
	)
)
func main() {
	ctx := context.Background()
	// TODO: It takes around one minute to detect the full edge of voltage change. Needs to tune the report period
	initOpenCensusCompare("opencensus-java-stats-demo-app", 1)

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

func RecordTemperatureHumidityCompare(ctx context.Context, pin int) {
	for range time.Tick(5 * time.Second) {
		defer logger.FinalizeLogger()
		// Uncomment/comment next line to suppress/increase verbosity of output
		logger.ChangePackageLogLevel("dht", logger.InfoLevel)

		sensorType := dht.DHT11
		// Read DHT11 sensor data from pin 4, retrying 10 times in case of failure.
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
		stats.Record(ctx, humidityMeasureCompare.M(float64(humidity)))
		if retried > 10 {
		}
	}
}

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
