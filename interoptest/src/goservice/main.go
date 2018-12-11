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
	"log"
	"net/http"
	"strings"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/gorilla/mux"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

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

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello, it is goservice " + message
	w.Write([]byte(message))
}

func main() {
	registerOcAgentExporter()
	registerJaegerExporter()
	r := mux.NewRouter()
	r.HandleFunc("/", sayHello)

	var handler http.Handler = r
	handler = &ochttp.Handler{ // add opencensus instrumentation
		Handler:     handler,
		Propagation: &tracecontext.HTTPFormat{}}

	if err := http.ListenAndServe(":10201", handler); err != nil {
		panic(err)
	}
}
