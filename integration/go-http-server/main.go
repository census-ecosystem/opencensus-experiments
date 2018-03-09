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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/google"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	pb "github.com/census-instrumentation/opencensus-experiments/integration/proto"
)

func echo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	inSpan := trace.FromContext(ctx)
	sCtx := inSpan.SpanContext()
	res := &pb.EchoResponse{
		TraceId:      []byte(sCtx.TraceID[:]),
		SpanId:       []byte(sCtx.SpanID[:]),
		TraceOptions: int32(sCtx.TraceOptions),
	}

	tagMap := tag.FromContext(ctx)
	// TODO: (@odeke-em) when https://github.com/census-instrumentation/opencensus-go/issues/521
	// is resolved, then we can retrieve tag keys and values
	if tagMap != nil {
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(res); err != nil {
		log.Fatalf("Failed to encode response from %q", r.RemoteAddr)
	}
}

type multiPropagationHandler struct {
	mux *http.ServeMux
}

func (mph *multiPropagationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var fh propagation.HTTPFormat
	switch p := q.Get("p"); p {
	default:
		http.Error(w, fmt.Sprintf("no such propagator %q", p), http.StatusBadRequest)
		return

	case "b3":
		fh = new(b3.HTTPFormat)
	case "google":
		fh = new(google.HTTPFormat)
	case "tracecontext":
		fh = new(tracecontext.HTTPFormat)
	}

	atRuntimeHandler := &ochttp.Handler{
		Handler:     mph.mux,
		Propagation: fh,
	}

	atRuntimeHandler.ServeHTTP(w, r)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(echo))

	addr := os.Getenv("OPENCENSUS_GO_HTTP_INTEGRATION_TEST_SERVER_ADDR")
	if addr == "" {
		addr = ":9900"
	}

	// At runtime we need to be able to multiplex on
	// the various propagators from this interop test.
	mph := &multiPropagationHandler{mux: mux}
	if err := http.ListenAndServe(addr, mph); err != nil {
		log.Fatalf("Go gRPC server failed to serve: %v", err)
	}
}
