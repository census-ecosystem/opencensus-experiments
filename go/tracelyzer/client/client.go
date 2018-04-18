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

package tracelyzer

import (
	"go.opencensus.io/trace"
	"sync"
	"reflect"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/tracelyzerpb"
	"google.golang.org/api/support/bundler"
	"log"
	"google.golang.org/grpc"
	"context"
	"github.com/gogo/protobuf/proto"
	"io"
	"github.com/golang/groupcache/consistenthash"
)

type Exporter struct {
	mu       sync.Mutex
	nodes    []string
	ring     *consistenthash.Map
	bundlers map[string]*bundler.Bundler
	opts     Options
}

type Options struct {
	Discovery   Discovery
	DialOptions []grpc.DialOption
	OnError  func(error)
}

var _ trace.Exporter = &Exporter{}

func NewExporter(o Options) *Exporter {
	e := &Exporter{
		bundlers: make(map[string]*bundler.Bundler),
		opts:     o,
	}
	var d Discovery
	if o.Discovery == nil {
		d = &NodeList{"localhost:9000"}
	} else {
		d = o.Discovery
	}
	e.resetNodes(d.GetNodes())
	go e.pollForUpdates(d.Poll())
	return e
}

func (e *Exporter) pollForUpdates(nodesCh chan []string) {
	for nodes := range nodesCh {
		e.mu.Lock()
		if !reflect.DeepEqual(e.nodes, nodes) {
			e.resetNodes(nodes)
		}
		e.mu.Unlock()
	}
}

func (e *Exporter) resetNodes(nodes []string) {
	e.ring = consistenthash.New(len(nodes), nil)
	e.ring.Add(nodes...)
	e.nodes = nodes
}

func (e *Exporter) ExportSpan(span *trace.SpanData) {
	e.mu.Lock()
	defer e.mu.Unlock()
	traceID := string(span.TraceID[:])
	n := e.ring.Get(traceID)
	buf, ok := e.bundlers[n]
	if !ok {
		buf = bundler.NewBundler(&tracelyzerpb.Span{}, func(bundle interface{}) {
			e.submitSpans(n, bundle.([]*tracelyzerpb.Span))
		})
		buf.BufferedByteLimit = 10 * 1024 * 1024
		e.bundlers[n] = buf
	}
	protoSpan := convertSpan(span)
	buf.Add(protoSpan, proto.Size(protoSpan))
}

func (e *Exporter) submitSpans(to string, spans []*tracelyzerpb.Span) {
	cc, err := grpc.Dial(to, e.opts.DialOptions...)
	if err != nil {
		e.reportError(err)
		return
	}
	defer cc.Close()
	client := tracelyzerpb.NewTracelyzerClient(cc)
	sender, err := client.SubmitSpan(context.TODO())
	if err != nil {
		e.reportError(err)
		return
	}
	var req tracelyzerpb.SubmitSpanRequest
	for _, span := range spans {
		req.Span = span
		sender.Send(&req)
	}
	_, err = sender.CloseAndRecv()
	if err != io.EOF {
		e.reportError(err)
	}
}

func (e *Exporter) reportError(err error) {
	if e.opts.OnError != nil {
		e.opts.OnError(err)
	} else {
		log.Printf("Error in tracelyzer exporter: %s", err)
	}
}

func (e *Exporter) Flush() {
	for _, b := range e.bundlers {
		b.Flush()
	}
}
