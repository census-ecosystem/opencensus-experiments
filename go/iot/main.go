package main

import (
	"github.com/mgutz/logxi/v1"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

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

const (
	high int = 1
	low  int = 0
)

func main() {
	r := raspi.NewAdaptor()
	myGPIO := gpio.NewDirectPinDriver(r, "11")
	led := gpio.NewLedDriver(r, "7")
	work := func() {
		// TODO: Since the sample period is 1 seconds, the worst delay would be 1 sec
		// Since this is a simple demo applciation, we could temporary ignore this part.
		gobot.Every(1*time.Second, func() {
			voltage, err := myGPIO.DigitalRead()
			if err != nil {
				log.Error("Error with Reading Voltage on the Raspberry Pi Pin 11")
			} else {
				if voltage == high {
					led.On()
				} else {
					led.Off()
				}
			}
		})
	}

	robot := gobot.NewRobot("PinVoltageCollection",
		[]gobot.Connection{r},
		[]gobot.Device{myGPIO, led},
		work,
	)

	robot.Start()
}
