package main

import (
	"github.com/huin/goserial"
	"io/ioutil"
	"strings"
	"fmt"
	"time"
	"encoding/json"
	"io"
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
const jsonStream = `
	{"Sensor": "Ed", "Time": 123}
	{"Sensor": "Sam", "Time": 456}
	{"Sensor": "Ed", "Time": 789}
	{"Sensor": "Sam", "Time": 10}
	{"Sensor": "Ed", "Time": 15}
`
type Message struct {
	Sensor string
	Time int
}
func main() {
	// Find the device that represents the arduino serial
	// connection.
	c := &goserial.Config{Name: findArduino(), Baud: 115200}
	s, _ := goserial.OpenPort(c)
	dec := json.NewDecoder(strings.NewReader(jsonStream))
	time.Sleep(3 * time.Second)
	var m Message
	for true{
		time.Sleep(time.Millisecond * 600)
		if err := dec.Decode(&m); err == io.EOF {
			fmt.Println(err.Error())
			continue
		} else if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("%s: %d\n", m.Sensor, m.Time)
	}
	s.Close()

}