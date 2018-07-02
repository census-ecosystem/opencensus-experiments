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
	"fmt"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/Protocol"
	"github.com/huin/goserial"
	"log"
	"time"
)

type Slave struct {
	// TODO: Use pointers?
	listeners []*OpenCensusBase
	// TODO: Maybe We could directly use ReaderWriter?
	reader *bufio.Reader
	sender *bufio.Writer
}

func (slave *Slave) notifyCensusToRegister(arguments *Protocol.Argument) {
	for _, census := range slave.listeners {
		err := census.InitOpenCensus(arguments)
		// TODO: Should I throw back the error or directly respond with error like this
		if err != nil {
			slave.respond(Protocol.FAIL, err.Error())
		} else {
			slave.respond(Protocol.OK, "Regist Successfully")
			log.Println("Regist Successfully!")
		}
	}
}

func (slave *Slave) notifyCensusToRecord(arguments *Protocol.Argument) {
	for _, census := range slave.listeners {
		err := census.Record(arguments)
		// TODO: Should I throw back the error or directly respond with error like this
		if err != nil {
			slave.respond(Protocol.FAIL, err.Error())
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
		time.Sleep(2 * time.Second)
		slave.reader = bufio.NewReader(s)
		slave.sender = bufio.NewWriter(s)
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
	slave.sender.Flush()
	slave.sender.Write([]byte(b))
	slave.sender.WriteByte('\n')
}

func (slave *Slave) Collect(period time.Duration) {
	// TODO: Default is that we regard every time from the arduino might be different measurement.
	for true {
		select {
		case <-time.After(period):
			input, isPrefix, err := slave.reader.ReadLine()
			fmt.Println(string(input))
			if err != nil {
				log.Println("Could Not Read the data from the Port")
				continue
			}
			if isPrefix == true {
				//TODO: The length of the json is bigger than the buffer size
				continue
			} else {
				var argument Protocol.Argument
				decodeErr := json.Unmarshal(input, &argument)
				if decodeErr != nil {
					log.Println(decodeErr)
				} else {
					switch argument.ArgumentType {
					case Protocol.REGISTRATION:
						slave.notifyCensusToRegister(&argument)
					case Protocol.RECORD:
						slave.notifyCensusToRecord(&argument)
					default:
						slave.respond(Protocol.FAIL, "Unsupported Argument Type")
					}
				}
			}
		}
	}
}
