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

package interoptestservice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/genproto"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/receiver"
	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

// ServerImpl is the type that handles RPCs for interop test service.
type ServerImpl struct {
	mu     sync.Mutex
	ln     net.Listener
	server *grpc.Server

	svc *ServiceImpl

	stopOnce         sync.Once
	startServerOnce  sync.Once
	startServiceOnce sync.Once
}

func NewServer(addr string) (*ServerImpl, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to address %q: error: %v", addr, err)
	}
	srv := &ServerImpl{ln: ln}
	return srv, nil
}

var (
	errAlreadyStarted = errors.New("already started")
	errAlreadyStopped = errors.New("already stopped")
)

// Start runs ServiceImpl.
func (srv *ServerImpl) Start(ctx context.Context, svc *ServiceImpl) error {
	if err := srv.registerService(svc); err != nil && err != errAlreadyStarted {
		return err
	}

	if err := srv.startGRPCServer(); err != nil && err != errAlreadyStarted {
		return err
	}

	return nil
}

func (srv *ServerImpl) registerService(svc *ServiceImpl) error {
	var err = errAlreadyStarted

	srv.startServiceOnce.Do(func() {
		srv.svc = svc
		grpcSrv := srv.grpcServer()
		interop.RegisterInteropTestServiceServer(grpcSrv, svc)
		err = nil
	})

	return err
}

func (srv *ServerImpl) grpcServer() *grpc.Server {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.server == nil {
		srv.server = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	}

	return srv.server
}

// Stop stops the underlying gRPC server and all the services running on it.
func (srv *ServerImpl) Stop() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var err = errAlreadyStopped
	srv.stopOnce.Do(func() {
		srv.server.GracefulStop()
		_ = srv.ln.Close()
	})
	return err
}

func (srv *ServerImpl) startGRPCServer() error {
	err := errAlreadyStarted
	srv.startServerOnce.Do(func() {
		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.server.Serve(srv.ln)
		}()

		// Our goal is to heuristically try running the server
		// and if it returns an error immediately, we reporter that.
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

// ServiceImpl is the implementation of InteropTestServiceServer.
type ServiceImpl struct {
	mu                 sync.Mutex
	registeredServices map[string][]*interop.Service
	sink               *receiver.TestCoordinatorSink
}

// NewService returns a new ServiceImpl with the given registered services.
func NewService(services map[string][]*interop.Service, sink *receiver.TestCoordinatorSink) *ServiceImpl {
	return &ServiceImpl{registeredServices: services, sink: sink}
}

func (s *ServiceImpl) SetRegisteredServices(services map[string][]*interop.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.registeredServices = services
}

func (s *ServiceImpl) Result(ctx context.Context, req *interop.InteropResultRequest) (*interop.InteropResultResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reps := &interop.InteropResultResponse{
		Id:     req.Id,
		Status: &interop.CommonResponseStatus{Status: interop.Status_SUCCESS},
		// TODO: store and return cached result
		Result: []*interop.TestResult{},
	}
	return reps, nil
}

// Runs the test asynchronously.
func (s *ServiceImpl) Run(ctx context.Context, req *interop.InteropRunRequest) (*interop.InteropRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := rand.Int63()
	// TODO: run all tests by sending out test execution requests
	verifySpans(s.sink.SpansPerNode)
	return &interop.InteropRunResponse{Id: id}, nil
}

func verifySpans(map[*commonpb.Node][]*tracepb.Span) {
	// TODO: implement this
}
