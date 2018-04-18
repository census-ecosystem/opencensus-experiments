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

package analyzes

import (
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/server/store"
	"go.opencensus.io/trace"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/tracelyzerpb"
	"go.opencensus.io/exporter/stackdriver"

	"github.com/golang/protobuf/ptypes"
	"os"
	example "go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"golang.org/x/time/rate"
	"go.opencensus.io/stats"
	"context"
	"time"
)

type allExporter interface {
	trace.Exporter
	view.Exporter
}

var exporters []allExporter

func init() {
	// TODO: make configurable
	projectID := os.Getenv("STACKDRIVER_PROJECT_ID")
	e, _ := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
	})
	exporters = []allExporter{e, &example.PrintExporter{}}
	for _, e := range exporters {
		view.RegisterExporter(e)
		trace.RegisterExporter(e)
	}
}

var rateLimit = rate.NewLimiter(10, 10)
var rateLimited = stats.Int64(
	"opencensus.io/trace/error-rate-limited",
	"Number of times we were rate limited on submitting error spans",
	stats.UnitNone)

func MaybeExport(t *store.Trace) {
	if hasError(t) {
		record(t, "error")
	}
	if rootSpanSlow(t) {
		record(t, "slow")
	}
}

// TODO: make configurable
var slowTime = 500 * time.Millisecond

func rootSpanSlow(t *store.Trace) bool {
	if r := t.Root(); r != nil {
		start, _ := ptypes.Timestamp(r.StartTime)
		end, _ := ptypes.Timestamp(r.EndTime)
		dur := end.Sub(start)
		if dur > slowTime {
			return true
		}
	}
	return false
}

func hasError(t *store.Trace) bool {
	for _, span := range t.Spans {
		if span.Status.Code != 0 {
			return true
		}
	}
	return false
}

func record(t *store.Trace, reason string) {
	if !rateLimit.Allow() {
		stats.Record(context.Background(), rateLimited.M(1))
	}
	for _, span := range t.Spans {
		data := convertToSpanData(t.TraceID, span)
		data.Attributes["tracelyzer"] = reason
		for _, exporter := range exporters {
			exporter.ExportSpan(data)
		}
	}
}

func convertToSpanData(traceID string, span *tracelyzerpb.Span) *trace.SpanData {
	var data trace.SpanData
	data.Name = span.Name.Value
	data.Status.Code = span.Status.Code
	data.Status.Message = span.Status.Message
	data.Attributes = make(map[string]interface{})
	copy(data.TraceID[:], []byte(traceID))
	copy(data.SpanID[:], []byte(span.SpanId))
	copy(data.ParentSpanID[:], []byte(span.ParentSpanId))
	data.StartTime, _ = ptypes.Timestamp(span.StartTime)
	data.EndTime, _ = ptypes.Timestamp(span.EndTime)

	// TODO: attributes, links, events, etc.

	return &data
}
