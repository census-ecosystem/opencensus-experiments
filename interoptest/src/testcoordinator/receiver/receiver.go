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

package receiver

import (
	"context"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/census-instrumentation/opencensus-service/receiver/opencensus"
)

// NewOCTraceReceiver creates a new OpenCensus Receiver at the given address. Also binds TraceServiceGrpc to
// the created gRPC server and starts Trace reception.
func NewOCTraceReceiver(addr string) (*opencensus.Receiver, *TestCoordinatorSink, error) {
	var receiver *opencensus.Receiver
	var sink *TestCoordinatorSink
	var err error
	if receiver, err = opencensus.New(addr); err == nil {
		sink = &TestCoordinatorSink{SpansPerNode: make(map[*commonpb.Node][]*tracepb.Span)}
		err = receiver.StartTraceReception(context.Background(), sink)
	}
	return receiver, sink, err
}
