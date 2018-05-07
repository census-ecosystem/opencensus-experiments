// Copyright 2018 Google Inc. All Rights Reserved.
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

// Command tracelyzer-run generates spans around execs.
package main

import (
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/client"
	"flag"
	"go.opencensus.io/trace"
	"context"
	"google.golang.org/grpc"
	"fmt"
	"os"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"net/http"
	"os/exec"
	"strings"
	example "go.opencensus.io/examples/exporter"
)

var (
	serverAddress string
	spanName      string
	parent       string
)

func main() {
	flag.Usage = func() {
		bin := "tracelyzer-submit"
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [command...]\n", bin)
		fmt.Fprintf(os.Stderr, "Example: %s -server localhost:9000 sleep 1\n", bin)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}
	flag.StringVar(&serverAddress, "server", "localhost:9000", "Server address")
	flag.StringVar(&spanName, "spanName", "Cmd.Run", "Span name")
	flag.StringVar(&parent, "parent", "", "Value of Trace-Parent header. If unset, a new trace is generated.")
	flag.Parse()


	exporter := tracelyzer.NewExporter(tracelyzer.Options{
		Discovery:   &tracelyzer.NodeList{serverAddress},
		DialOptions: []grpc.DialOption{grpc.WithInsecure()},
	})
	trace.RegisterExporter(exporter)
	trace.RegisterExporter(&example.PrintExporter{})
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	var span *trace.Span
	var headerFormat tracecontext.HTTPFormat
	req := &http.Request{
		Header: make(http.Header),
	}

	if parent != "" {
		req.Header.Add("Trace-Parent", parent)
		sc, ok := headerFormat.SpanContextFromRequest(req)
		if !ok {
			fmt.Println("Invalid Trace-Parent header:", parent)
			os.Exit(1)
		}
		_, span = trace.StartSpanWithRemoteParent(context.Background(), spanName, sc)
	} else {
		_, span = trace.StartSpan(context.Background(), spanName)
		headerFormat.SpanContextToRequest(span.SpanContext(), req)
		fmt.Printf("Trace-Parent: %s\n", req.Header["Trace-Parent"][0])
	}

	if flag.NArg() > 0 {
		span.AddAttributes(trace.StringAttribute("command", strings.Join(flag.Args(), " ")))
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		err := cmd.Run()
		if err != nil {
			span.SetStatus(trace.Status{Code: 2, Message: err.Error()})
		} else {
			span.SetStatus(trace.Status{Code: 0})
		}
	}

	span.End()
	exporter.Flush()
}
