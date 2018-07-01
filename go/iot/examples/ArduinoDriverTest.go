package main

import (
	"github.com/census-ecosystem/opencensus-experiments/go/iot/openCensus"
	"github.com/huin/goserial"
	"io/ioutil"
	"strings"
	"time"
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
	census.Initialize()
	slave.Initialize(c)
	slave.Subscribe(census)
	slave.Collect(2 * time.Second)
}
