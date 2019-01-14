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

package validator

import (
	"errors"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/tracestate"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
)

// SimpleSpan is a tree-like type that preserves the parent-child relationship among spans and
// holds the propagation information (trace id, span id, etc.).
type SimpleSpan struct {
	children map[trace.SpanID]*SimpleSpan

	traceID    trace.TraceID
	spanID     trace.SpanID
	tracestate tracestate.Tracestate
}

var (
	errOrphanSpan         = errors.New("found orphan span")
	errAlreadyExists      = errors.New("found two spans with the same span ID")
	errDuplicatedRootSpan = errors.New("found duplicated root span for the same trace")
)

// ReconstructTraces tries to reconstruct traces from the given spans. If the spans are valid, it will return the root
// spans for each trace. If there are broken traces, an error for each broken one will be returned instead.
func ReconstructTraces(spans ...*tracepb.Span) (map[trace.TraceID]*SimpleSpan, map[trace.TraceID]error) {
	dict := groupSpansByTraceID(spans)
	roots := map[trace.TraceID]*SimpleSpan{}
	errs := map[trace.TraceID]error{}
	processedSpans := map[trace.SpanID]*SimpleSpan{} // cache processed spans for faster loop-up
outerLoop:
	for tid, spans := range dict { // iterate each trace
		for len(spans) > 0 {
			// The order of spans in the list are non-deterministic,
			// so we need to keep iterating over the span list until either:
			// 1. all spans are processed if the spans can form a valid trace;
			// 2. return an error if the spans cannot form a valid trace.
			size := len(spans)
			for i, span := range spans {
				ok, err := processSpan(tid, roots, processedSpans, span)
				if err != nil {
					errs[tid] = err
					deleteIfExists(roots, tid)
					continue outerLoop
				}
				if ok { // if processed, remove the span from the to-be-processed slice.
					spans = removeFromSpanList(i, spans)
					break // process one span at an iteration since the size of slice changed.
				}
			}
			// After processing, check if we processed a span. If not, either there's no root span, or there're spans
			// whose parent doesn't exist. Either cases indicate there's an orphan spans.
			if size == len(spans) {
				errs[tid] = errOrphanSpan
				deleteIfExists(roots, tid)
				continue outerLoop
			}
		}
	}
	return roots, errs
}

func groupSpansByTraceID(spans []*tracepb.Span) map[trace.TraceID][]*tracepb.Span {
	dict := map[trace.TraceID][]*tracepb.Span{}
	for _, span := range spans {
		traceID := ToTraceID(span.TraceId)
		dict[traceID] = append(dict[traceID], span)
	}
	return dict
}

func processSpan(tid trace.TraceID, roots map[trace.TraceID]*SimpleSpan, processedSpans map[trace.SpanID]*SimpleSpan, span *tracepb.Span) (bool, error) {
	if processedSpans[ToSpanID(span.SpanId)] != nil {
		return false, errAlreadyExists
	}
	if isRoot(span) { // root span
		if roots[tid] == nil {
			root := spanToSimpleSpan(span)
			roots[tid] = root
			processedSpans[root.spanID] = root
			return true, nil
		} else { // each trace can only have one root span
			return false, errDuplicatedRootSpan
		}
	} else { // leaf span
		psID := ToSpanID(span.ParentSpanId)
		parent := processedSpans[psID] // check if we already processed its parent
		if parent != nil {
			child := spanToSimpleSpan(span)
			parent.children[child.spanID] = child
			processedSpans[child.spanID] = child
			return true, nil
		}
		return false, nil
	}
}

// ToTraceID creates a Trace ID from the given byte array.
func ToTraceID(bytes []byte) trace.TraceID {
	var bytesCopy [16]byte
	copy(bytesCopy[:], bytes[:16])
	return trace.TraceID(bytesCopy)
}

// ToSpanID creates a SpanID ID from the given byte array.
func ToSpanID(bytes []byte) trace.SpanID {
	var bytesCopy [8]byte
	copy(bytesCopy[:], bytes[:8])
	return trace.SpanID(bytesCopy)
}

func toTracestate(tspb *tracepb.Span_Tracestate) tracestate.Tracestate {
	entries := make([]tracestate.Entry, len(tspb.Entries))
	for _, e := range tspb.Entries {
		entry := tracestate.Entry{Key: e.Key, Value: e.Value}
		entries = append(entries, entry)
	}
	ts, _ := tracestate.New(nil, entries...)
	return *ts
}

func spanToSimpleSpan(span *tracepb.Span) *SimpleSpan {
	ss := &SimpleSpan{
		children: make(map[trace.SpanID]*SimpleSpan),
		traceID:  ToTraceID(span.TraceId),
		spanID:   ToSpanID(span.SpanId),
	}
	if span.Tracestate != nil && span.Tracestate.Entries != nil {
		ss.tracestate = toTracestate(span.Tracestate)
	}
	return ss
}

func isRoot(span *tracepb.Span) bool {
	return span.ParentSpanId == nil || len(span.ParentSpanId) == 0
}

func removeFromSpanList(i int, spans []*tracepb.Span) []*tracepb.Span {
	if len(spans) == 0 || i < 0 || i >= len(spans) {
		return spans
	}
	spans[i] = spans[len(spans)-1]
	return spans[:len(spans)-1]
}

func deleteIfExists(roots map[trace.TraceID]*SimpleSpan, id trace.TraceID) {
	_, ok := roots[id]
	if ok {
		delete(roots, id)
	}
}
