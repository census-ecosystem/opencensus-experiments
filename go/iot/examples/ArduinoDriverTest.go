package main

import (
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/arduino"
	"fmt"
)

func main() {
	a := raspi.NewAdaptor()
	adc := arduino.NewArduinoDriver(a)

	work := func() {
		adc.TransferAndWait('a')
		adc.TransferAndWait(10)
		a, err := adc.TransferAndWait(17)
		b, err := adc.TransferAndWait(33)
		c, err := adc.TransferAndWait(42)
		d, err := adc.TransferAndWait(0)
		fmt.Printf("A: %d, B: %d, C: %d, D: %d\n", a, b, c, d)
		if err == nil{

		}
		select {}
	}

	robot := gobot.NewRobot("ArduinoDriver",
		[]gobot.Connection{a},
		[]gobot.Device{adc},
		work,
	)

	robot.Start()
}
