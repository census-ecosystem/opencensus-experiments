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

package registrationservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/genproto"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

// Handler is the type that handles registration requests.
type Handler struct {
	mu     sync.Mutex
	ln     net.Listener
	server *grpc.Server

	Receiver *RegistrationReceiver

	stopOnce              sync.Once
	startServerOnce       sync.Once
	startRegistrationOnce sync.Once
}

var (
	errAlreadyStarted = errors.New("already started")
	errAlreadyStopped = errors.New("already stopped")
)

// New just creates the registration services.
func New(addr string) (*Handler, error) {
	// TODO: consider using options.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind to address %q: error: %v", addr, err)
	}
	h := &Handler{ln: ln}

	return h, nil
}

func (h *Handler) registerRegistrationReceiver() error {
	var err = errAlreadyStarted

	h.startRegistrationOnce.Do(func() {
		h.Receiver = &RegistrationReceiver{RegisteredServices: make(map[string][]*interop.Service)}
		srv := h.grpcServer()
		interop.RegisterRegistrationServiceServer(srv, h.Receiver)
		err = nil
	})

	return err
}

func (h *Handler) grpcServer() *grpc.Server {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.server == nil {
		h.server = grpc.NewServer()
	}

	return h.server
}

// Start runs the registration service.
func (h *Handler) Start(ctx context.Context) error {
	if err := h.registerRegistrationReceiver(); err != nil && err != errAlreadyStarted {
		return err
	}

	if err := h.startGRPCServer(); err != nil && err != errAlreadyStarted {
		return nil
	}

	// At this point we've successfully started all the services/receivers.
	// Add other start routines here.
	return nil
}

// Stop stops the underlying gRPC server and all the services running on it.
func (h *Handler) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var err = errAlreadyStopped
	h.stopOnce.Do(func() {
		h.server.GracefulStop()
		_ = h.ln.Close()
	})
	return err
}

func (h *Handler) startGRPCServer() error {
	err := errAlreadyStarted
	h.startServerOnce.Do(func() {
		errChan := make(chan error, 1)
		go func() {
			errChan <- h.server.Serve(h.ln)
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
