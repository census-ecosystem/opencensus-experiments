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

package testservice

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/goservice/genproto"
	"google.golang.org/grpc"
)

// GrpcReceiver is the type that handles test requests over GRPC.
type GrpcReceiver struct {
	mu     sync.Mutex
	ln     net.Listener
	server *grpc.Server

	receiver *GrpcTestReceiver

	stopOnce              sync.Once
	startServerOnce       sync.Once
	startRegistrationOnce sync.Once
}

// New just creates the test services for request over GRPC.
func New(addr string) (*GrpcReceiver, error) {
	// TODO: consider using options.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind to address %q: error: %v", addr, err)
	}
	h := &GrpcReceiver{ln: ln}

	return h, nil
}

func (gr *GrpcReceiver) registerTestReceiver() error {
	var err = errAlreadyStarted

	gr.startRegistrationOnce.Do(func() {
		gr.receiver = &GrpcTestReceiver{registeredServices: make(map[string][]*interop.TestRequest)}
		srv := gr.grpcServer()
		interop.RegisterTestExecutionServiceServer(srv, gr.receiver)
		err = nil
	})

	return err
}

func (gr *GrpcReceiver) grpcServer() *grpc.Server {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	if gr.server == nil {
		gr.server = grpc.NewServer()
	}

	return gr.server
}

// Start runs the test service.
func (gr *GrpcReceiver) Start(ctx context.Context) error {
	if err := gr.registerTestReceiver(); err != nil && err != errAlreadyStarted {
		return err
	}

	if err := gr.startGRPCServer(); err != nil && err != errAlreadyStarted {
		return err
	}

	// At this point we've successfully started all the services/receivers.
	// Add other start routines here.
	return nil
}

// Stop stops the underlying gRPC server and all the services running on it.
func (gr *GrpcReceiver) Stop() error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	var err = errAlreadyStopped
	gr.stopOnce.Do(func() {
		gr.server.GracefulStop()
		_ = gr.ln.Close()
	})
	return err
}

func (gr *GrpcReceiver) startGRPCServer() error {
	err := errAlreadyStarted
	gr.startServerOnce.Do(func() {
		errChan := make(chan error, 1)
		go func() {
			errChan <- gr.server.Serve(gr.ln)
		}()

		select {
		case serr := <-errChan:
			err = serr

		case <-time.After(1 * time.Second):
			// No error otherwise returned in the period of 1s.
			// We can assume that the serve is at least running.
			err = nil
		}
	})
	return err
}

// GrpcTestReceiver is the type used to handle test requests.
type GrpcTestReceiver struct {
	registeredServices map[string][]*interop.TestRequest
}

// Test is the gRPC method that handles test requests.
func (gtr *GrpcTestReceiver) Test(_ context.Context, req *interop.TestRequest) (*interop.TestResponse, error) {
	// TODO: add servicing test request.
	return &interop.TestResponse{Id: req.GetId()}, nil
}
