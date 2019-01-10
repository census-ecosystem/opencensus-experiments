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
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace/propagation"
	"goservice/genproto"
	"io/ioutil"
	"net/http"
)

type httpServer struct {
	mu     sync.Mutex
	server *http.Server

	stopOnce              sync.Once
	startServerOnce       sync.Once
	startRegistrationOnce sync.Once
}

// HttpReceiver is the type that handles test requests over HTTP.
type HttpReceiver struct {
	b3Server httpServer
	tcServer httpServer
}

func newHttpHandler(propagation propagation.HTTPFormat) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/test/request", httpTestRequestHandler)

	var h http.Handler = r
	h = &ochttp.Handler{ // add opencensus instrumentation
		Handler:     h,
		Propagation: propagation}
	return h
}

func newB3Server(b3Addr string) httpServer {
	return httpServer{server: &http.Server{Addr: b3Addr, Handler: newHttpHandler(&b3.HTTPFormat{})}}
}

func newTcServer(tcAddr string) httpServer {
	return httpServer{server: &http.Server{Addr: tcAddr, Handler: newHttpHandler(&tracecontext.HTTPFormat{})}}
}

// NewHttpReceiver just creates the test HTTP Server that services Test request over HTTP
// It creates two servers, one for B3 propagation and other for TraceContext propagation.
func NewHttpReceiver(b3Addr, tcAddr string) *HttpReceiver {
	return &HttpReceiver{b3Server: newB3Server(b3Addr), tcServer: newTcServer(tcAddr)}
}

func httpTestRequestHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}
	var rp *RequestProcessor
	testRequest := interop.TestRequest{}
	if err := proto.Unmarshal(data, &testRequest); err == nil {
		testResp, _ := rp.getInstance().process(req.Context(), &testRequest)
		data, err := proto.Marshal(testResp)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshalling response %s", err.Error()), http.StatusInternalServerError)
		} else {
			w.Write(data)
		}
	} else {
		http.Error(w, "error parsing request", http.StatusBadRequest)
	}
}

// B3Start starts the underlying HTTP Server using B3 Propagation.
func (hr *HttpReceiver) B3Start(ctx context.Context) error {
	return hr.b3Server.start(ctx)
}

// B3Stop stops the underlying HTTP Server using B3 Propagation.
func (hr *HttpReceiver) B3Stop() error {
	return hr.b3Server.stop()
}

// TcStart starts the underlying HTTP Server using TraceContext Propagation.
func (hr *HttpReceiver) TcStart(ctx context.Context) error {
	return hr.tcServer.start(ctx)
}

// TcStop stops the underlying HTTP Server using TraceContext Propagation.
func (hr *HttpReceiver) TcStop() error {
	return hr.tcServer.stop()
}

func (hs *httpServer) start(ctx context.Context) error {

	if err := hs.startHttpServer(); err != nil && err != errAlreadyStarted {
		return err
	}

	// At this point we've successfully started all the services/receivers.
	// Add other start routines here.
	return nil
}

func (hs *httpServer) stop() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	var err = errAlreadyStopped
	hs.stopOnce.Do(func() {
		hs.server.Shutdown(nil)
	})
	return err
}

func (hs *httpServer) startHttpServer() error {
	err := errAlreadyStarted
	hs.startServerOnce.Do(func() {
		errChan := make(chan error, 1)
		go func() {
			errChan <- hs.server.ListenAndServe()
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
