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

package receiver

import (
	"context"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/census-instrumentation/opencensus-service/data"
	"github.com/census-instrumentation/opencensus-service/receiver"
)

// TestCoordinatorSink is a struct that implements receiver.TraceReceiverSink. It can accept
// TraceData as well as expose stored TraceData for validation.
type TestCoordinatorSink struct {
	SpansPerNode map[*commonpb.Node][]*tracepb.Span
}

// ReceiveTraceData receives the span data in the protobuf format, groups them by Node and stores them.
func (tcs TestCoordinatorSink) ReceiveTraceData(ctx context.Context, td data.TraceData) (*receiver.TraceReceiverAcknowledgement, error) {
	node := td.Node
	spans := td.Spans

	tcs.SpansPerNode[node] = append(tcs.SpansPerNode[node], spans...)

	ack := &receiver.TraceReceiverAcknowledgement{
		SavedSpans: uint64(len(td.Spans)),
	}
	return ack, nil
}
