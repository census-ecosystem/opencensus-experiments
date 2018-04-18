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
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/tracelyzerpb"
	"go.opencensus.io/trace"
	"time"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func convertSpan(span *trace.SpanData) *tracelyzerpb.Span {
	return &tracelyzerpb.Span{
		Name:         convertString(span.Name),
		TraceId:      span.TraceID[:],
		SpanId:       span.SpanID[:],
		ParentSpanId: span.ParentSpanID[:],
		Kind:         convertKind(span.SpanKind),
		TimeEvents:   convertTimeEvents(span),
		Links:        convertLinks(span),
		Attributes:   convertAttributes(span),
		StartTime:    convertTime(span.StartTime),
		EndTime:      convertTime(span.EndTime),
		Status:       convertStatus(span.Status),
	}
}
func convertStatus(status trace.Status) *tracelyzerpb.Status {
	return &tracelyzerpb.Status{
		Code: status.Code,
		Message: status.Message,
	}
}

func convertTime(v time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: v.Unix(),
		Nanos:   int32(v.Nanosecond()),
	}
}

func convertAttributes(data *trace.SpanData) *tracelyzerpb.Span_Attributes {
	return nil
}

func convertLinks(data *trace.SpanData) *tracelyzerpb.Span_Links {
	return nil
}

func convertTimeEvents(data *trace.SpanData) *tracelyzerpb.Span_TimeEvents {
	//TODO: implement
	return nil
}

func convertString(s string) *tracelyzerpb.TruncatableString {
	//TODO: truncate if necessary
	return &tracelyzerpb.TruncatableString{
		Value:              s,
		TruncatedByteCount: 0,
	}
}

func convertKind(kind int) tracelyzerpb.Span_SpanKind {
	switch kind {
	case trace.SpanKindServer:
		return tracelyzerpb.Span_SERVER
	case trace.SpanKindClient:
		return tracelyzerpb.Span_CLIENT
	default:
		return tracelyzerpb.Span_SPAN_KIND_UNSPECIFIED
	}
}
