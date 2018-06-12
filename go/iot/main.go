package main

import (
	"gobot.io/x/gobot/platforms/raspi"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot"
	"fmt"
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

func main() {
	fmt.Println("HelloWorld!")
}

func process_video(){
	r := raspi.NewAdaptor()
	button := gpio.NewDirectPinDriver(r, "11")
	led := gpio.NewLedDriver(r, "7")

	work := func() {
	}

	robot := gobot.NewRobot("buttonBot",
		[]gobot.Connection{r},
		[]gobot.Device{button, led},
		work,
	)

	robot.Start()
}