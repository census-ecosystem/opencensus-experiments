// Copyright 2018, OpenCensus Authors
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

package openCensus

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"time"
	"fmt"

	"github.com/census-ecosystem/opencensus-experiments/go/iot/Protocol"
	"github.com/huin/goserial"
)

const (
	SETUPDURATION = 2
)
type Slave struct {
	listeners []*OpenCensusBase
	reader    *bufio.Reader
	// The serial library doesn't support bufio.NewWriter(io.ReadWriteCloser)
	sender io.ReadWriteCloser
}

func (slave *Slave) notifyCensusToRecord(arguments *Protocol.MeasureArgument) {
	for _, census := range slave.listeners {
		code, err := census.Record(arguments)
		if err != nil {
			slave.respond(code, err.Error())
		} else {
			slave.respond(Protocol.OK, "Record Successfully")
			log.Println("Record Successfull!")
		}
	}
}

func (slave *Slave) Subscribe(listener OpenCensusBase) {
	slave.listeners = append(slave.listeners, &listener)
}

func (slave *Slave) Initialize(config *goserial.Config) error {
	if s, err := goserial.OpenPort(config); err == nil {
		// It should wait for some time to let the serial initialization
		time.Sleep(SETUPDURATION * time.Second)
		slave.reader = bufio.NewReader(s)
		slave.sender = s
		return nil
	} else {
		return err
	}
}

func (slave *Slave) respond(code int, info string) {
	response := Protocol.Response{code, info}
	b, err := json.Marshal(response)
	if err != nil {
		log.Fatal("Could not encode the project\n")
	}
	slave.sender.Write([]byte(b))
	slave.sender.Write([]byte("\n"))
	log.Printf("Send Response: Code %d Info %s\n", response.Code, response.Info)
}

func (slave *Slave) Collect(period time.Duration) {
	go func() {
		for range time.Tick(period) {
			input, isPrefix, err := slave.reader.ReadLine()
			fmt.Println(string(input))
			if err != nil {
				log.Println("Could Not Read the data from the Port because %s", err.Error())
				continue
			}
			if isPrefix == true {
				//TODO: The length of the json is bigger than the buffer size
				continue
			} else {
				var argument Protocol.MeasureArgument
				decodeErr := json.Unmarshal(input, &argument)
				if decodeErr != nil {
					log.Println(decodeErr)
				} else {
					slave.notifyCensusToRecord(&argument)
				}
			}
		}
	}()
}
