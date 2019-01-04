// Copyright 2019, OpenCensus Authors
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

// cmd is the main launcher of test coordinator.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/interoptestservice"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/receiver"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/registrationservice"
	"github.com/census-instrumentation/opencensus-service/receiver/opencensus"
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
	// 1. Parse service addresses from command line.
	regSrvAddr, receiverAddr, interopSrvAddr := parseFlags()

	// 2. Create and start registration server.
	regSrv := startRegistrationService(regSrvAddr)
	defer regSrv.Stop()

	// 3. Create and start ocagent receiver.
	receiver, sink := startOCTraceReceiver(receiverAddr)
	defer receiver.Stop()

	// 4. After test services registered themselves, create interop service and start the server.
	// Once interopSrv started, controller can send test requests and retrieve test results via RPCs.
	interopSrv := startInteropTestService(interopSrvAddr, regSrv, sink)
	defer interopSrv.Stop()

	signalsChan := make(chan os.Signal)
	signal.Notify(signalsChan, os.Interrupt)

	// Wait for the closing signal
	<-signalsChan
}

func parseFlags() (regSrvAddr, receiverAddr, interopSrvAddr string) {
	regSrvAddrPtr := flag.String("registration_service_address", "localhost:10002", "Address of the registration service")
	recieverAddrPtr := flag.String("oc_receiver_address", "localhost:10001", "Address of OC Agent trace receiver")
	interopSrvAddrPtr := flag.String("interop_service_address", "localhost:10003", "Address of the interop test service")
	flag.Parse()
	return *regSrvAddrPtr, *recieverAddrPtr, *interopSrvAddrPtr
}

func startRegistrationService(addr string) *registrationservice.Handler {
	regSrv, err := registrationservice.New(addr)
	if err != nil {
		log.Errorf("error creating registration server: %v", err)
		os.Exit(-1)
	}
	if err = regSrv.Start(context.Background()); err != nil {
		log.Errorf("error starting registration server: %v", err)
		os.Exit(-1)
	}
	return regSrv
}

func startOCTraceReceiver(addr string) (*opencensus.Receiver, *receiver.TestCoordinatorSink) {
	receiver, sink, err := receiver.NewOCTraceReceiver(addr)
	if err != nil {
		log.Errorf("error creating ocagent receiver: %v", err)
		os.Exit(-1)
	}
	return receiver, sink
}

func startInteropTestService(addr string, regSrv *registrationservice.Handler, sink *receiver.TestCoordinatorSink) *interoptestservice.ServerImpl {
	interopService := interoptestservice.NewService(regSrv.Receiver.RegisteredServices, sink)
	interopSrv, err := interoptestservice.NewServer(addr)
	if err != nil {
		log.Errorf("error creating interop test server: %v", err)
		os.Exit(-1)
	}
	if err = interopSrv.Start(context.Background(), interopService); err != nil {
		log.Errorf("error starting interop test server: %v", err)
		os.Exit(-1)
	}
	return interopSrv
}
