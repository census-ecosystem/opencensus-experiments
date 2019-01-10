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
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/testdata"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/testexecutionservice"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/validator"
	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
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

// NewServer returns a new unstarted gRPC interop server.
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
	testSuites         []*interop.TestRequest
	results            []*interop.TestResult
}

// NewService returns a new ServiceImpl with the given registered services.
func NewService(services map[string][]*interop.Service, sink *receiver.TestCoordinatorSink) *ServiceImpl {
	return &ServiceImpl{registeredServices: services, sink: sink, testSuites: testdata.LoadTestSuites()}
}

// SetRegisteredServices sets the registered services to the given one.
func (s *ServiceImpl) SetRegisteredServices(services map[string][]*interop.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.registeredServices = services
}

// Result returns all the test results.
func (s *ServiceImpl) Result(ctx context.Context, req *interop.InteropResultRequest) (*interop.InteropResultResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp := &interop.InteropResultResponse{
		Id:     req.Id,
		Status: &interop.CommonResponseStatus{Status: interop.Status_SUCCESS},
		Result: s.results,
	}
	return resp, nil
}

// Run runs the test asynchronously.
func (s *ServiceImpl) Run(ctx context.Context, req *interop.InteropRunRequest) (*interop.InteropRunResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := rand.Int63()

outer:
	for _, req := range s.testSuites {
		svc := req.ServiceHops[0].Service // always send request to the first hop to initiate tests
		req.ServiceHops = req.ServiceHops[1:]

		sender, _ := testexecutionservice.NewUnstartedSender(true, req.Id, req.Name, fmt.Sprintf("%s:%d", svc.Host, svc.Port), req.ServiceHops)
		resp, err := sender.Start()

		if err != nil {
			s.results = append(s.results, getFailedResult(req.Id, req.Name, req.ServiceHops, nil))
			continue outer
		}

		for _, status := range resp.Status {
			if status.Status == interop.Status_FAILURE {
				s.results = append(s.results, getFailedResult(req.Id, req.Name, req.ServiceHops, resp.Status))
				continue outer
			}
		}
	}

	time.Sleep(60 * time.Second) // wait until all micro-services exported their spans
	results := verifyAllSpans(s.sink.SpansPerNode)
	s.sink.SpansPerNode = map[*commonpb.Node][]*tracepb.Span{} // flush verified spans
	s.results = append(s.results, results...)

	return &interop.InteropRunResponse{Id: id}, nil
}

func getFailedResult(id int64, reqName string, hops []*interop.ServiceHop, details []*interop.CommonResponseStatus) *interop.TestResult {
	return &interop.TestResult{
		Id:          id,
		Name:        reqName,
		Status:      &interop.CommonResponseStatus{Status: interop.Status_FAILURE},
		ServiceHops: hops,
		Details:     details,
	}
}

func verifyAllSpans(spansPerNode map[*commonpb.Node][]*tracepb.Span) []*interop.TestResult {
	var exportedSpans []*tracepb.Span
	for _, spans := range spansPerNode {
		exportedSpans = append(exportedSpans, spans...)
	}
	reqIDByTraceID := groupReqIDsByTraceID(exportedSpans)
	traces, errs := validator.ReconstructTraces(exportedSpans...)
	return generateResultForEachReq(reqIDByTraceID, traces, errs)
}

const reqIDKey = "reqId"

func groupReqIDsByTraceID(spans []*tracepb.Span) map[trace.TraceID]int64 {
	reqIDByTraceID := map[trace.TraceID]int64{}
	for _, span := range spans {
		reqID := span.Attributes.AttributeMap[reqIDKey].GetIntValue()
		if reqID == 0 {
			continue
		}
		reqIDByTraceID[validator.ToTraceID(span.TraceId)] = reqID
	}
	return reqIDByTraceID
}

func generateResultForEachReq(reqIDByTraceID map[trace.TraceID]int64, traces map[trace.TraceID]*validator.SimpleSpan, errs map[trace.TraceID]error) []*interop.TestResult {
	results := []*interop.TestResult{}
	for traceID := range traces {
		result := &interop.TestResult{
			Id:     reqIDByTraceID[traceID],
			Status: &interop.CommonResponseStatus{Status: interop.Status_SUCCESS},
			// TODO: add req name and service hops?
		}
		results = append(results, result)
	}
	for traceID, err := range errs {
		errMsg := fmt.Sprintf("Unable to reconstruct traces %s, error :%s", string(traceID[:16]), err.Error())
		result := &interop.TestResult{
			Id:     reqIDByTraceID[traceID],
			Status: &interop.CommonResponseStatus{Status: interop.Status_FAILURE, Error: errMsg},
			// TODO: add req name and service hops?
		}
		results = append(results, result)
	}
	return results
}
