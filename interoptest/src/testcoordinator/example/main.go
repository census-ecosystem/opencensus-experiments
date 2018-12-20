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

// Example contains a program that demonstrates how all components work together.
package main

import (
	"context"
	"os"
	"time"

	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/interoptestservice"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/receiver"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/registrationservice"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

func main() {
	// 1. Create and start registration server.
	regSrv, err := registrationservice.New("localhost:10002")
	if err != nil {
		log.Errorf("error creating registration server: %v", err)
		os.Exit(-1)
	}
	if err = regSrv.Start(context.Background()); err != nil {
		log.Errorf("error starting registration server: %v", err)
		os.Exit(-1)
	}

	// 2. Create and start ocagent receiver.
	_, sink, err := receiver.NewOCTraceReceiver("localhost:10001")
	if err != nil {
		log.Errorf("error creating ocagent receiver: %v", err)
		os.Exit(-1)
	}

	// 3. After test services registered themselves, create interop service and start the server.
	interopService := interoptestservice.NewService(regSrv.Receiver.RegisteredServices, sink)
	interopSrv, err := interoptestservice.NewServer("localhost:10003")
	if err != nil {
		log.Errorf("error creating interop test server: %v", err)
		os.Exit(-1)
	}
	if err = interopSrv.Start(context.Background(), interopService); err != nil {
		log.Errorf("error starting interop test server: %v", err)
		os.Exit(-1)
	}

	// Once interopSrv started, controller can send test requests and retrieve test results via RPCs.
}
