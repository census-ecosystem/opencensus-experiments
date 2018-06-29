package driver

import (
	"bufio"
	"encoding/json"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/openCensus"
	"github.com/huin/goserial"
	"log"
	"time"
)

type Slave struct {
	// TODO: Use pointers?
	listeners []openCensus.OpenCensusBase
	// TODO: Maybe We could directly use ReaderWriter?
	reader *bufio.Reader
	sender *bufio.Writer
}

func (slave *Slave) notifyCensusToRegister(arguments *openCensus.RegisterArgument) {
	for _, census := range slave.listeners {
		err := census.InitOpenCensus(arguments)
		// TODO: Should I throw back the error or directly respond with error like this
		if err != nil {
			slave.respond(err)
		} else {
			slave.respond(err)
		}
	}
}

func (slave *Slave) notifyCensusToRecord(arguments *openCensus.RecordArgument) {
	for _, census := range slave.listeners {
		err := census.Record(arguments)
		// TODO: Should I throw back the error or directly respond with error like this
		if err != nil {
			slave.respond(err)
		} else {
			slave.respond(err)
		}
	}
}

func (slave *Slave) Subscribe(listener openCensus.OpenCensusBase) {
	slave.listeners = append(slave.listeners, listener)
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
	slave.sender.Flush()
	slave.sender.Write([]byte(err.Error()))
	// TODO: use \r\n or \n???
	slave.sender.WriteByte('\n')
}

func (slave *Slave) Collect(period time.Duration) {
	// TODO: Default is that we regard every time from the arduino might be different measurement.
	for true {
		select {
		case <-time.After(period):
			input, isPrefix, err := slave.reader.ReadLine()
			if err != nil {
				slave.respond(err)
				continue
			}
			if isPrefix == true {
				//TODO: The length of the json is bigger than the buffer size
				continue
			} else {
				var m openCensus.Argument
				decodeErr := json.Unmarshal(input, &m)
				if decodeErr != nil {
					log.Println(err)
				} else {
					// TODO:
				}
			}
		}
	}
}
