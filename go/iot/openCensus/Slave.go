package openCensus

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/Protocol"
	"github.com/huin/goserial"
	"github.com/pkg/errors"
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
			slave.respond(err)
		} else {
			//slave.respond(err)
			// TODO: Modify here
			log.Println("Registrate Successfully!")
		}
	}
}

func (slave *Slave) notifyCensusToRecord(arguments *Protocol.Argument) {
	for _, census := range slave.listeners {
		err := census.Record(arguments)
		// TODO: Should I throw back the error or directly respond with error like this
		if err != nil {
			slave.respond(err)
		} else {
			//slave.respond(err)
			log.Println("Record Successfully!")
		}
	}
}

func (slave *Slave) Subscribe(listener OpenCensusBase) {
	slave.listeners = append(slave.listeners, &listener)
}

func (slave *Slave) Initialize(config *goserial.Config) error {
	if s, err := goserial.OpenPort(config); err == nil {
		time.Sleep(2 * time.Second)
		slave.reader = bufio.NewReader(s)
		slave.sender = bufio.NewWriter(s)
		return nil
	} else {
		return err
	}
}

func (slave *Slave) respond(err error) {
	log.Fatal(err)
	/*
		slave.sender.Flush()
		slave.sender.Write([]byte(err.Error()))
		// TODO: use \r\n or \n???
		slave.sender.WriteByte('\n')
	*/
}

func (slave *Slave) Collect(period time.Duration) {
	// TODO: Default is that we regard every time from the arduino might be different measurement.
	for true {
		select {
		case <-time.After(period):
			input, isPrefix, err := slave.reader.ReadLine()
			fmt.Println(string(input))
			if err != nil {
				slave.respond(err)
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
						slave.respond(errors.Errorf("ArgumentType Not Supported\n!"))
					}
				}
			}
		}
	}
}
