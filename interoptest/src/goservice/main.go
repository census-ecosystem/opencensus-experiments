// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"
	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
	"goservice/genproto"
	"goservice/testservice"
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

func registerJaegerExporter() {

	// Register the Jaeger exporter to be able to retrieve
	// the collected spans.
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint: "http://traceui:14268",
		Process: jaeger.Process{
			ServiceName: "goservice",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)
}

func registerOcAgentExporter() {
	oce, err := ocagent.NewExporter(ocagent.WithInsecure(), ocagent.WithAddress("ocagent:55678"))
	if err != nil {
		//log.Fatalf("Failed to create ocagent-exporter: %v", err)
	}
	trace.RegisterExporter(oce)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
}

func main() {
	// For debugging use JaegerExporter.
	// registerJaegerExporter()
	registerOcAgentExporter()
	grpcServer, err := testservice.NewGRPCReciever(fmt.Sprintf(":%d", interop.ServicePort_GO_GRPC_BINARY_PROPAGATION_PORT))
	if err != nil {
		log.Errorf("error creating grpc server: %v", err)
		os.Exit(-1)
	}
	b3Addr := fmt.Sprintf(":%d", interop.ServicePort_GO_HTTP_B3_PROPAGATION_PORT)
	tcAddr := fmt.Sprintf(":%d", interop.ServicePort_GO_HTTP_TRACECONTEXT_PROPAGATION_PORT)
	httpServer := testservice.NewHttpReceiver(b3Addr, tcAddr)

	ctx := context.Background()
	err = grpcServer.Start(ctx)
	if err != nil {
		log.Errorf("error starting grpc server: %v", err)
		os.Exit(-1)
	}

	err = httpServer.B3Start(ctx)
	if err != nil {
		log.Errorf("error starting http b3 server: %v", err)
		os.Exit(-1)
	}

	err = httpServer.TcStart(ctx)
	if err != nil {
		log.Errorf("error starting http tracecontext server: %v", err)
		os.Exit(-1)
	}

	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	<-gracefulStop
	log.Println("goservice: shutdown hook, Listening signals...")

	if grpcServer != nil {
		grpcServer.Stop()
	}
	if httpServer != nil {
		httpServer.B3Stop()
		httpServer.TcStop()
	}
	log.Println("goservice terminated. wait for 2 seconds")
	time.Sleep(2 * time.Second)
}
