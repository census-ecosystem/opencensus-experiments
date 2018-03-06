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
	"log"
	"net"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/plugin/ocgrpc"

	pb "github.com/census-instrumentation/opencensus-experiments/integration/proto"
)

type server int

var _ pb.EchoServiceServer = (*server)(nil)

func (s *server) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	// 1. Retrieve the TraceID
	// 2. Retrieve the SpanID
	// 3. Retrieve the Tags
	// 4. Retrieve the TraceOptions
	inSpan := trace.FromContext(ctx)
	sCtx := inSpan.SpanContext()
	res := &pb.EchoResponse{
		TraceId:      []byte(sCtx.TraceID[:]),
		SpanId:       []byte(sCtx.SpanID[:]),
		TraceOptions: int32(sCtx.TraceOptions),
	}

	tagMap := tag.FromContext(ctx)
	// TODO: File a Go issue for tagMap key inspection
	// as the underlying map is unexposed.
	if tagMap != nil {
	}

	return res, nil
}

const addr = ":9800"

func main() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Go gRPC server net.Listen: %v", err)
	}

	srv := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	pb.RegisterEchoServiceServer(srv, new(server))
	if err := srv.Serve(ln); err != nil {
		log.Fatalf("Go gRPC server failed to serve: %v", err)
	}
}
