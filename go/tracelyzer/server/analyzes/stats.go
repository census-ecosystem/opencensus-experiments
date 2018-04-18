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
	"time"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/tracelyzerpb"
	"github.com/golang/protobuf/ptypes"
	"go.opencensus.io/tag"
	"go.opencensus.io/stats"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/server/store"
	"context"
	"fmt"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/plugin/ocgrpc"
)

var (
	spanDurationPerTrace = stats.Float64("opencensus.io/trace/span-duration", "Span durationn", stats.UnitMilliseconds)
	spanCountPerTrace    = stats.Int64("opencensus.io/trace/span-count", "Span count per trace", stats.UnitNone)
	spanName, _          = tag.NewKey("span_name")
)

type spanStats struct {
	count    int64
	totalDur time.Duration
}

type key struct {
	name     string
	spanKind tracelyzerpb.Span_SpanKind
}

func init() {
	view.Register(
		&view.View{
			Measure:     spanCountPerTrace,
			Aggregation: ocgrpc.DefaultMessageCountDistribution,
		},
		&view.View{
			Measure:     spanDurationPerTrace,
			Aggregation: ocgrpc.DefaultMillisecondsDistribution,
		},
	)
}

func RecordStats(trace *store.Trace) {
	spansByName := make(map[key]spanStats)
	for _, span := range trace.Spans {
		end, err := ptypes.Timestamp(span.EndTime)
		if err != nil {
			continue
		}
		start, err := ptypes.Timestamp(span.StartTime)
		if err != nil {
			continue
		}
		k := key{name: span.Name.Value, spanKind: span.Kind}
		st := spansByName[k]
		st.count++
		st.totalDur += end.Sub(start)
		spansByName[k] = st
	}
	ctx := context.Background()
	for k, spanStats := range spansByName {
		ctx, _ = tag.New(ctx, tag.Insert(spanName, k.describe()))
		stats.Record(ctx,
			spanCountPerTrace.M(spanStats.count),
			spanDurationPerTrace.M(float64(spanStats.totalDur/time.Millisecond)))
	}
}

func (k key) describe() string {
	switch k.spanKind {
	case tracelyzerpb.Span_CLIENT:
		return fmt.Sprintf("Client.%s", k.name)
	case tracelyzerpb.Span_SERVER:
		return fmt.Sprintf("Server.%s", k.name)
	default:
		return k.name
	}
}
