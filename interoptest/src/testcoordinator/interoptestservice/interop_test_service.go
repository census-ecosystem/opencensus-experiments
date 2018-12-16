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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
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

var _ http.Handler = (*ServiceImpl)(nil)

// ServeHTTP allows ServiceImpl to handle HTTP requests.
func (s *ServiceImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		// For options, unconditionally respond with a 200 and send CORS headers.
		// Without properly responding to OPTIONS and without CORS headers, browsers
		// won't be able to use this handler.
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.WriteHeader(200)
		return
	}

	// Handle routing.
	switch r.Method {
	case "GET":
		s.handleHTTPGET(w, r)
		return

	case "POST":
		s.handleHTTPPOST(w, r)
		return

	default:
		http.Error(w, "Unhandled HTTP Method: "+r.Method+" only accepting POST and GET", http.StatusMethodNotAllowed)
		return
	}
}

func deserializeJSON(blob []byte, save interface{}) error {
	return json.Unmarshal(blob, save)
}

func serializeJSON(src interface{}) ([]byte, error) {
	return json.Marshal(src)
}

// readTillEOFAndDeserializeJSON reads the entire body out of rc and then closes it.
// If it encounters an error, it will return it immediately.
// After successfully reading the body, it then JSON unmarshals to save.
func readTillEOFAndDeserializeJSON(rc io.ReadCloser, save interface{}) error {
	// We are always receiving an interop.InteropResultRequest
	blob, err := ioutil.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		return err
	}
	return deserializeJSON(blob, save)
}

const resultPathPrefix = "/result/"

func (s *ServiceImpl) handleHTTPGET(w http.ResponseWriter, r *http.Request) {
	// Expecting a request path of: "/result/:id"
	var path string
	if r.URL != nil {
		path = r.URL.Path
	}

	if len(path) <= len(resultPathPrefix) {
		http.Error(w, "Expected path of the form: /result/:id", http.StatusBadRequest)
		return
	}

	strId := strings.TrimPrefix(path, resultPathPrefix)
	if strId == "" || strId == "/" {
		http.Error(w, "Expected path of the form: /result/:id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: actually look up the available tests by their IDs
	req := &interop.InteropResultRequest{Id: id}
	res, err := s.Result(r.Context(), req)
	if err != nil {
		// TODO: perhaps multiplex on NotFound and other sentinel errors.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blob, _ := serializeJSON(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(blob)
}

func (s *ServiceImpl) handleHTTPPOST(w http.ResponseWriter, r *http.Request) {
	var path string
	if r.URL != nil {
		path = r.URL.Path
	}

	ctx := r.Context()
	var res interface{}
	var err error

	switch path {
	case "/run", "/run/":
		inrreq := new(interop.InteropRunRequest)
		if err := readTillEOFAndDeserializeJSON(r.Body, inrreq); err != nil {
			http.Error(w, "Failed to JSON unmarshal interop.InteropRunRequest: "+err.Error(), http.StatusBadRequest)
			return
		}
		res, err = s.Run(ctx, inrreq)

	case "/result", "/result/":
		inrreq := new(interop.InteropResultRequest)
		if err := readTillEOFAndDeserializeJSON(r.Body, inrreq); err != nil {
			http.Error(w, "Failed to JSON unmarshal interop.InteropResultRequest: "+err.Error(), http.StatusBadRequest)
			return
		}
		res, err = s.Result(ctx, inrreq)

	default:
		http.Error(w, "Unmatched route: "+path+"\nOnly accepting /result and /run", http.StatusNotFound)
		return
	}

	if err != nil {
		// TODO: Perhap return a structured error e.g. {"error": <ERROR_MESSAGE>}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Otherwise all clear to return the response.
	blob, _ := serializeJSON(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(blob)
}
