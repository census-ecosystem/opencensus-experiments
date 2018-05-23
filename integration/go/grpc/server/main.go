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
	"os"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"

	pb "github.com/census-ecosystem/opencensus-experiments/integration/proto"
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
		TagsBlob:     tag.Encode(tag.FromContext(ctx)),
		TraceOptions: int32(sCtx.TraceOptions),
	}

	return res, nil
}

func main() {
	addr := os.Getenv("OPENCENSUS_GO_GRPC_INTEGRATION_TEST_SERVER_ADDR")
	if addr == "" {
		addr = ":9800"
	}

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
