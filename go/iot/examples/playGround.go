package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/huin/goserial"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type M struct {
	Sensor string
	Time   int64
}

func findArduino2() string {
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
	c := &goserial.Config{Name: findArduino2(), Baud: 9600}
	s, err := goserial.OpenPort(c)
	if err != nil {

	}
	time.Sleep(2 * time.Second)
	reader := bufio.NewReader(s)
	//sender := bufio.NewWriter(s)

	for true {
		select {
		case <-time.After(500 * time.Millisecond):
			input, isPrefix, err := reader.ReadLine()
			fmt.Println(input)
			if err != nil {
				log.Println(err)
				continue
			}
			if isPrefix == true {
				//TODO: The length of the json is bigger than the buffer size
				continue
			} else {
				var argument M
				decodeErr := json.Unmarshal(input, &argument)
				if decodeErr != nil {
					log.Println(err)
				} else {
					log.Println("HellowWorld!")
				}
			}
		}
	}

}
