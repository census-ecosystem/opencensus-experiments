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

package main

import (
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/tracelyzerpb"
	"google.golang.org/grpc"
	"net"
	"fmt"
	"log"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/server/store"
	"time"
	"net/http"
	"go.opencensus.io/zpages"
	"io"
	"go.opencensus.io/trace"
	"go.opencensus.io/plugin/ocgrpc"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/server/analyzes"
)

var port int32 = 9000


func main() {
	// set up z-pages
	http.Handle("/debug", http.StripPrefix("/debug/", zpages.Handler))
	go http.ListenAndServe(":8080", zpages.Handler)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	server := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{
		StartOptions: trace.StartOptions{ Sampler: trace.NeverSample() },
	}))

	traces := make(chan *store.Trace, 1024)
	st := store.NewStore(10*time.Second, traces)
	go processTraces(traces)
	tracelyzerpb.RegisterTracelyzerServer(server, &trServer{
		store: st,
	})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve(lis)
}

type trServer struct {
	store *store.Store
}

func (t *trServer) SubmitSpan(ss tracelyzerpb.Tracelyzer_SubmitSpanServer) error {
	for {
		req, err := ss.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		traceID := string(req.Span.TraceId)
		req.Span.TraceId = nil
		t.store.PutSpan(traceID, req.Span)
	}
}

func processTraces(traces chan *store.Trace) {
	for t := range traces {
		analyzes.RecordStats(t)
		analyzes.MaybeExport(t)
	}
}
