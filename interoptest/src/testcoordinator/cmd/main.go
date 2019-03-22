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

	interoppb "github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/genproto"

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
	// Currently micro services are registered statically and this server performs no-op.
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
	regSrvAddrPtr := flag.String("registration_service_address", "0.0.0.0:10002", "Address of the registration service")
	recieverAddrPtr := flag.String("oc_receiver_address", "0.0.0.0:10001", "Address of OC Agent trace receiver")
	interopSrvAddrPtr := flag.String("interop_service_address", "0.0.0.0:10003", "Address of the interop test service")
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

var (
	grpcBinarySpec = &interoppb.Spec{
		Transport:   interoppb.Spec_GRPC,
		Propagation: interoppb.Spec_BINARY_FORMAT_PROPAGATION,
	}

	httpB3Spec = &interoppb.Spec{
		Transport:   interoppb.Spec_HTTP,
		Propagation: interoppb.Spec_B3_FORMAT_PROPAGATION,
	}

	httpTCSpec = &interoppb.Spec{
		Transport:   interoppb.Spec_HTTP,
		Propagation: interoppb.Spec_TRACE_CONTEXT_FORMAT_PROPAGATION,
	}

	microSvcs = map[string][]*interoppb.Service{
		"java": []*interoppb.Service{
			&interoppb.Service{
				Name: "java:grpc:binary",
				Port: int32(interoppb.ServicePort_JAVA_GRPC_BINARY_PROPAGATION_PORT),
				Host: "javaservice",
				Spec: grpcBinarySpec,
			},
			&interoppb.Service{
				Name: "java:http:b3",
				Port: int32(interoppb.ServicePort_JAVA_HTTP_B3_PROPAGATION_PORT),
				Host: "javaservice",
				Spec: httpB3Spec,
			},
			&interoppb.Service{
				Name: "java:http:tc",
				Port: int32(interoppb.ServicePort_JAVA_HTTP_TRACECONTEXT_PROPAGATION_PORT),
				Host: "javaservice",
				Spec: httpTCSpec,
			},
		},
		"go": []*interoppb.Service{
			&interoppb.Service{
				Name: "go:grpc:binary",
				Port: int32(interoppb.ServicePort_GO_GRPC_BINARY_PROPAGATION_PORT),
				Host: "goservice",
				Spec: grpcBinarySpec,
			},
			&interoppb.Service{
				Name: "go:http:b3",
				Port: int32(interoppb.ServicePort_GO_HTTP_B3_PROPAGATION_PORT),
				Host: "goservice",
				Spec: httpB3Spec,
			},
			&interoppb.Service{
				Name: "go:http:tc",
				Port: int32(interoppb.ServicePort_GO_HTTP_TRACECONTEXT_PROPAGATION_PORT),
				Host: "goservice",
				Spec: httpTCSpec,
			},
		},
		"python": []*interoppb.Service{
			&interoppb.Service{
				Name: "python:grpc:binary",
				Port: int32(interoppb.ServicePort_PYTHON_GRPC_BINARY_PROPAGATION_PORT),
				Host: "pythonservice",
				Spec: grpcBinarySpec,
			},
			&interoppb.Service{
				Name: "python:http:b3",
				Port: int32(interoppb.ServicePort_PYTHON_HTTP_B3_PROPAGATION_PORT),
				Host: "pythonservice",
				Spec: httpB3Spec,
			},
			&interoppb.Service{
				Name: "python:http:tc",
				Port: int32(interoppb.ServicePort_PYTHON_HTTP_TRACECONTEXT_PROPAGATION_PORT),
				Host: "pythonservice",
				Spec: httpTCSpec,
			},
		},
		"node": []*interoppb.Service{
			&interoppb.Service{
				Name: "node:grpc:binary",
				Port: int32(interoppb.ServicePort_NODEJS_GRPC_BINARY_PROPAGATION_PORT),
				Host: "nodeservice",
				Spec: grpcBinarySpec,
			},
			&interoppb.Service{
				Name: "node:http:b3",
				Port: int32(interoppb.ServicePort_NODEJS_HTTP_B3_PROPAGATION_PORT),
				Host: "nodeservice",
				Spec: httpB3Spec,
			},
			&interoppb.Service{
				Name: "node:http:tc",
				Port: int32(interoppb.ServicePort_NODEJS_HTTP_TRACECONTEXT_PROPAGATION_PORT),
				Host: "nodeservice",
				Spec: httpTCSpec,
			},
		},
	}
)

func startInteropTestService(addr string, regSrv *registrationservice.Handler, sink *receiver.TestCoordinatorSink) *interoptestservice.ServerImpl {
	// Statically register the micro services for now.
	// TODO: allow micro services to be reigstered dynamically
	interopService := interoptestservice.NewService(microSvcs, sink)

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
