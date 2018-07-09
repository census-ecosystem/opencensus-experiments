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

// Program iot uploads sensor data including temperature, humidity, sound and light strength to monitoring backend by
// using the OpenCensus framework.
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/d2r2/go-dht"
	"github.com/d2r2/go-logger"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	// view to see the sound strength distribution.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	viewSoundDist = &view.View{
		Name:        "opencensus.io/views/sound_strength_distribution",
		Description: "sound strength distribution over time",
		Measure:     soundStrengthDistMeasure,
		Aggregation: view.Distribution(0, 2, 4, 8, 16, 32, 64, 128),
	}

	// view to see the sound strength instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	viewSoundLast = &view.View{
		Name:        "opencensus.io/views/sound_strength_instant",
		Description: "sound strength instantly over time",
		Measure:     soundStrengthLastMeasure,
		Aggregation: view.LastValue(),
	}

	// view to see the light strength instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	viewLight = &view.View{
		Name:        "opencensus.io/views/light_strength_instant",
		Description: "voltage level on GPIO over time",
		Measure:     lightStrengthMeasure,
		Aggregation: view.LastValue(),
	}

	// view to see the humidity instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	viewHumidity = &view.View{
		Name:        "opencensus.io/views/humidity_instant",
		Description: "humidity_over time",
		TagKeys:     []tag.Key{sensorKey},
		Measure:     humidityMeasure,
		Aggregation: view.LastValue(),
	}

	// view to see the temperature instantly.
	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	viewTemperature = &view.View{
		Name:        "opencensus.io/views/temperature_instant",
		Description: "temperature over time",
		TagKeys:     []tag.Key{sensorKey},
		Measure:     temperatureMeasure,
		Aggregation: view.LastValue(),
	}

	// Apply two kinds of aggregation type to the same metric in order to see the difference.
	soundStrengthDistMeasure = stats.Int64("opencensus.io/measure/sound_strength_svl_mp1_7c3c_dist", "strength of sound", stats.UnitDimensionless)
	soundStrengthLastMeasure = stats.Int64("opencensus.io/measure/sound_strength_svl_mp1_7c3c_last", "strength of sound", stats.UnitDimensionless)
	lightStrengthMeasure     = stats.Int64("opencensus.io/measure/light_strength_svl_mp1_7c3c", "strength of light", stats.UnitDimensionless)
	humidityMeasure          = stats.Float64("opencensus.io/measure/humidity_svl_mp1_7c3c", "humidity", stats.UnitDimensionless)
	temperatureMeasure       = stats.Float64("opencensus.io/measure/temperature_svl_mp1_7c3c", "temperature", stats.UnitDimensionless)

	soundSamplePeriod       = 50 * time.Millisecond
	temperatureSamplePeriod = 5 * time.Second

	sensorKey tag.Key

	lg = logger.NewPackageLogger("main",
		logger.DebugLevel,
		// logger.InfoLevel,
	)
)

// The board would connect to two DHT11 temperature sensors with the GPIO4 and GPIO17.
// Communicate with ADS1015S based on the I2C and connect the sound
// and light sensor to the A0, A1 channel on the ADC module.
func main() {
	ctx := context.Background()
	projectId := os.Getenv("PROJECTID")
	if projectId == "" {
		log.Fatal("Cannot detect PROJECTID in the system environment.\n")
	} else {
		log.Printf("Project Id is set to be %s\n", projectId)
	}

	var err error
	sensorKey, err = tag.NewKey("opencensus.io/keys/sensor")
	if err != nil {
		log.Fatal(err)
	}

	initOpenCensus(projectId, 1)
	// Create a new go thread to record the temperature and humidity
	go RecordTemperatureHumidity(ctx, 4)
	go RecordTemperatureHumidity(ctx, 17)

	board := raspi.NewAdaptor()
	ads1015 := i2c.NewADS1015Driver(board)
	soundSensor := aio.NewGroveSoundSensorDriver(ads1015, "0")
	lightSensor := aio.NewGroveLightSensorDriver(ads1015, "1")

	work := func() {
		gobot.Every(soundSamplePeriod, func() {
			// Since the sample shares the same pin, it cannot be done concurrently.
			recordSound(ctx, soundSensor)
			recordLight(ctx, lightSensor)
		})
	}

	robot := gobot.NewRobot("sensorDataCollection",
		[]gobot.Connection{board},
		[]gobot.Device{ads1015},
		work,
	)
	robot.Start()
}

// Record the sound strength based on two kinds of aggregation.
// One is distribution, the other is the lastValue.
func recordSound(ctx context.Context, soundSensor *aio.GroveSoundSensorDriver) {
	soundStrength, soundErr := readSound(soundSensor)
	if soundErr != nil {
		log.Fatalf("Could not read value from sound sensors\n")
	} else {
		stats.Record(ctx, soundStrengthDistMeasure.M(int64(soundStrength)))
		stats.Record(ctx, soundStrengthLastMeasure.M(int64(soundStrength)))
		//log.Printf("Sound Strength: %d\n", soundStrength)
	}
}

// Record the light strength.
func recordLight(ctx context.Context, lightSensor *aio.GroveLightSensorDriver) {
	lightStrength, lightErr := lightSensor.Read()
	if lightErr != nil {
		log.Fatalf("Could not read value from light sensors\n")
	} else {
		stats.Record(ctx, lightStrengthMeasure.M(int64(lightStrength)))
		//log.Printf("Light Strength: %d\n", lightStrength)
	}
}

// Sample 50 sound strength data in a period.
// Calculate the maximum and minimum value and return their difference
func readSound(sensor *aio.GroveSoundSensorDriver) (int, error) {
	min := math.MaxInt32
	max := math.MinInt32
	for i := 0; i < 50; i++ {
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

// For every five seconds, record the temperature and humidity sensor data.
// Print the collected data on the console.
func RecordTemperatureHumidity(ctx context.Context, pin int) {
	ctx, err := tag.New(ctx,
		tag.Insert(sensorKey, fmt.Sprintf("Sensor :%d", pin)),
	)
	if err != nil {
		log.Fatal(err)
	}
	for range time.Tick(temperatureSamplePeriod) {
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
			lg.Fatal(err)
		}
		if temperature > 0 && humidity > 0 && err != nil && retried > 0 {
		}
		// print temperature and humidity
		lg.Infof("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)",
			sensorType, temperature, humidity, retried)
		stats.Record(ctx, temperatureMeasure.M(float64(temperature)))
		stats.Record(ctx, humidityMeasure.M(float64(humidity)))
	}
}

// Initialize the openCensus framework.
// If there is anything wrong with the registration, directly throw a fatal error.
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

	// Set reporting period to report data based on the given reportPeriod.
	view.SetReportingPeriod(time.Second * time.Duration(reportPeriod))

	viewList := []*view.View{viewSoundDist, viewSoundLast, viewLight, viewHumidity, viewTemperature}

	for _, viewToRegister := range viewList {
		if err := view.Register(viewToRegister); err != nil {
			log.Fatalf("Cannot subscribe to the view: %v", err)
		}
	}

}
