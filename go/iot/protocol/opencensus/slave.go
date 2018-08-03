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

package opencensus

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol/parser"
	"github.com/huin/goserial"
)

const (
	// TODO: We could make it configurable in the future
	SETUPDURATION = 2
)

type Slave struct {
	listeners []*OpenCensusBase
	reader    *bufio.Reader
	// The serial library doesn't support bufio.NewWriter(io.ReadWriteCloser)
	sender   io.ReadWriteCloser
	myParser parser.Parser
}

func (slave *Slave) notifyCensusToRecord(arguments *protocol.MeasureArgument) {
	for _, census := range slave.listeners {
		response := census.Record(arguments)
		slave.respond(response)
	}
}

func (slave *Slave) Subscribe(listener OpenCensusBase) {
	slave.listeners = append(slave.listeners, &listener)
}

func (slave *Slave) Initialize(config *goserial.Config, parser parser.Parser) error {
	if s, err := goserial.OpenPort(config); err == nil {
		// It should wait for some time to initialize the arduino end
		time.Sleep(SETUPDURATION * time.Second)
		slave.reader = bufio.NewReader(s)
		slave.sender = s
		slave.myParser = parser
		return nil
	} else {
		return err
	}
}

func (slave *Slave) respond(response *protocol.Response) {
	b, err := json.Marshal(response)
	if err != nil {
		log.Fatal("Could not encode the project because", err)
	}
	slave.sender.Write([]byte(b))
	// Every response should be ended with the character '\n'
	// The info part in the response should not contain this character
	slave.sender.Write([]byte("\n"))
}

func (slave *Slave) Collect(period time.Duration) {
	go func() {
		for range time.Tick(period) {
			input, isPrefix, err := slave.reader.ReadLine()
			//fmt.Println(string(input))
			if err != nil {
				log.Printf("Could not read the data from the port because %s", err.Error())
				continue
			}
			if isPrefix == true {
				//TODO: The length of the json is bigger than the buffer size
				continue
			} else {
				output, decodeErr := slave.myParser.Parse(input)
				if decodeErr != nil {
					// If we don't respond here, there would deadlock between the arduino and Pi.
					response := protocol.Response{protocol.FAIL, "Fail to parse the message because " + decodeErr.Error()}
					slave.respond(&response)
				} else {
					slave.notifyCensusToRecord(&output)
				}
			}
		}
	}()
}
