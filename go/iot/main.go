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
	"math"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/d2r2/go-dht"
	"github.com/d2r2/go-logger"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	high = 1
)

var (
	soundStrengthMeasure = stats.Int64("my.org/measure/sound_strength_svl_mp1_7c3c", "strength of sound", stats.UnitDimensionless)
	lightStrengthMeasure = stats.Int64("my.org/measure/light_strength_svl_mp1_7c3c", "strength of light", stats.UnitDimensionless)
	humidityMeasure      = stats.Float64("my.org/measure/humidity_svl_mp1_7c3c", "humidity", stats.UnitDimensionless)
	temperatureMeasure   = stats.Float64("my.org/measure/temperature_svl_mp1_7c3c", "temperature", stats.UnitDimensionless)

	lg = logger.NewPackageLogger("main",
		logger.DebugLevel,
		// logger.InfoLevel,
	)
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
	initOpenCensus("opencensus-java-stats-demo-app", 1)

	go recordTemperatureHumidity(ctx)
	board := raspi.NewAdaptor()
	ads1015 := i2c.NewADS1015Driver(board)
	soundSensor := aio.NewGroveSoundSensorDriver(ads1015, "0")
	lightSensor := aio.NewGroveLightSensorDriver(ads1015, "1")

	work := func() {
		gobot.Every(1*time.Second, func() {
			recordSound(ctx, soundSensor)
			recordLight(ctx, lightSensor)
		})
	}

	robot := gobot.NewRobot("sensorDataCollection",
		[]gobot.Connection{board},
		[]gobot.Device{ads1015, soundSensor, lightSensor},
		work,
	)
	robot.Start()
}

func recordSound(ctx context.Context, soundSensor *aio.GroveSoundSensorDriver) {
	soundStrength, soundErr := readSound(soundSensor)
	if soundErr != nil {
		log.Fatalf("Could not read value from sound sensors\n")
	} else {
		stats.Record(ctx, soundStrengthMeasure.M(int64(soundStrength)))
	}
}

func recordLight(ctx context.Context, lightSensor *aio.GroveLightSensorDriver) {
	lightStrength, lightErr := lightSensor.Read()
	if lightErr != nil {
		log.Fatalf("Could not read value from light sensors\n")
	} else {
		stats.Record(ctx, lightStrengthMeasure.M(int64(lightStrength)))
	}
}

func readSound(sensor *aio.GroveSoundSensorDriver) (int, error) {
	min := math.MaxInt32
	max := math.MinInt32
	for i := 0; i < 100; i++ {
		strength, err := sensor.Read()
		if err != nil {
			log.Fatalf("Couldn't read data from the sensor\n")
		} else {
			if strength > max {
				max = strength
			}
			if strength < min {
				min = strength
			}
		}
	}
	return max - min, nil
}
func recordTemperatureHumidity(ctx context.Context) {
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
			dht.ReadDHTxxWithRetry(sensorType, 4, false, 50)
		if err != nil {
			lg.Fatal(err)
		}
		// print temperature and humidity
		//lg.Infof("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)",
		//	sensorType, temperature, humidity, retried)
		stats.Record(ctx, temperatureMeasure.M(float64(temperature)))
		stats.Record(ctx, humidityMeasure.M(float64(humidity)))
		if retried > 10 {
		}
	}
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

	// Create view to see the sound strength instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Register(&view.View{
		Name:        "my.org/views/sound_strength_distribution",
		Description: "sound strength over time",
		Measure:     soundStrengthMeasure,
		Aggregation: view.Distribution(0, 300, 350, 400, 450, 500),
	}); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}

	// Create view to see the light strength instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Register(&view.View{
		Name:        "my.org/views/light_strength_instant",
		Description: "voltage level on GPIO over time",
		Measure:     lightStrengthMeasure,
		Aggregation: view.LastValue(),
	}); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}

	// Create view to see the humidity instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Register(&view.View{
		Name:        "my.org/views/humidity_instant",
		Description: "humidity_over time",
		Measure:     humidityMeasure,
		Aggregation: view.LastValue(),
	}); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}

	// Create view to see the temperature instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := view.Register(&view.View{
		Name:        "my.org/views/temperature_instant",
		Description: "temperature over time",
		Measure:     temperatureMeasure,
		Aggregation: view.LastValue(),
	}); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v", err)
	}
}
