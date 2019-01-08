// Copyright 2019, OpenCensus Authors
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

package testservice

import (
	"fmt"
	"go.opencensus.io/plugin/ocgrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"goservice/genproto"
)

// GRPCSender is the type that handles test requests over gRPC.
type GRPCSender struct {
}

// NewGRPCSender creates a sender that can send Test request over gRPC.
func NewGRPCSender() *GRPCSender {
	return &GRPCSender{}
}

// Send sends gRPC request to a server specified by serviceHop.
func (gs *GRPCSender) Send(ctx context.Context, serviceHop interop.ServiceHop, req *interop.TestRequest) (*interop.TestResponse, error) {
	address := fmt.Sprintf("%s:%d", serviceHop.Service.Host, serviceHop.Service.Port)
	conn, err := grpc.Dial(address, grpc.WithStatsHandler(&ocgrpc.ClientHandler{}), grpc.WithInsecure())
	if err != nil {
		// TODO: log error
		// log.Fatalf("Cannot connect: %v", err)
	}
	defer conn.Close()
	c := interop.NewTestExecutionServiceClient(conn)

	// Contact the server and print out its response.
	r, err := c.Test(context.Background(), req)
	if err != nil {
		// TODO: log error
		// log.Printf("Could not test: %v", err)
	}
	return r, err
}
