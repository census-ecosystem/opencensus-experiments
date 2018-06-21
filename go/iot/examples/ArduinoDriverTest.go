package main

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/arduino"
	"time"
)

func main() {
	a := raspi.NewAdaptor()
	adc := arduino.NewArduinoDriver(a)

	work := func() {
		gobot.Every(1*time.Second, func() {
			var test string = "Hello World!\n"
			err := adc.Write([]byte(test))
			fmt.Println("A0", err)
		})
	}

	robot := gobot.NewRobot("ArduinoDriver",
		[]gobot.Connection{a},
		[]gobot.Device{adc},
		work,
	)

	robot.Start()
}
