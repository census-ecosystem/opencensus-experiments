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

package testexecutionservice

import (
	"context"
	"errors"
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/genproto"
	"sync"

	"google.golang.org/grpc"

	"go.opencensus.io/tag"
)

// Sender is the type that stores necessary information for making test requests, and sends
// test execution request to each test server.
type Sender struct {
	mu         sync.RWMutex
	startOnces []sync.Once

	canDialInsecure    bool

	// The order of reqIds, reqNames and serverAddrs must match.
	reqIds             []int64
	reqNames           []string
	serverAddrs        []string
	registeredServices map[string][]*interop.Service
	tagsForServices    map[string][]*tag.Tag
}

var (
	errAlreadyStarted = errors.New("already started")
	errSizeNotMatch   = errors.New("sizes do not match")
)

// NewUnstartedSender just creates a new Sender.
// TODO: consider using options.
func NewUnstartedSender(
	canDialInsecure bool,
	reqIds []int64,
	reqNames []string,
	serverAddrs []string,
	registeredServices map[string][]*interop.Service,
	tagsForServices map[string][]*tag.Tag) (*Sender, error) {
	if len(reqIds) != len(reqNames) || len(reqIds) != len(serverAddrs) || len(reqIds) != len(registeredServices) {
		return nil, errSizeNotMatch
	}
	startOnces := make([]sync.Once, len(reqIds))
	for i := range reqIds {
		startOnces[i] = sync.Once{}
	}
	s := &Sender{
		canDialInsecure:    canDialInsecure,
		reqIds:             reqIds,
		reqNames:           reqNames,
		serverAddrs:        serverAddrs,
		registeredServices: registeredServices,
		tagsForServices:    tagsForServices,
	}
	return s, nil
}

// Start transforms each request id, request name and Services into a TestRequest.
// Then sends each TestRequest to the corresponding server, and returns the list of responses
// and errors and for each request.
func (s *Sender) Start() ([]*interop.TestResponse, []error) {
	var resps []*interop.TestResponse
	var errs []error
	for i, so := range s.startOnces {
		var resp *interop.TestResponse
		err := errAlreadyStarted
		so.Do(func() {
			s.mu.Lock()
			defer s.mu.Unlock()

			addr := s.serverAddrs[i]
			if cc, err := s.dialToServer(addr); err == nil {
				resp, err = s.send(cc, s.reqIds[i], s.reqNames[i])
			}
		})
		resps = append(resps, resp)
		errs = append(errs, err)
	}
	return resps, errs
}

// TODO: send HTTP TestRequest
func (s *Sender) send(cc *grpc.ClientConn, reqId int64, reqName string) (*interop.TestResponse, error) {
	services := s.registeredServices[reqName]
	var hops []*interop.ServiceHop
	for _, service := range services {
		hops = append(hops, &interop.ServiceHop{
			Service: service,
			Tags:    toTagsProto(s.tagsForServices[service.Name]),
		})
	}
	req := &interop.TestRequest{
		Id:          reqId,
		Name:        reqName,
		ServiceHops: hops,
	}

	testSvcClient := interop.NewTestExecutionServiceClient(cc)
	return testSvcClient.Test(context.Background(), req)
}

func (s *Sender) dialToServer(addr string) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption
	if s.canDialInsecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	return grpc.Dial(addr, dialOpts...)
}

func toTagsProto(tags []*tag.Tag) []*interop.Tag {
	var tagsProto []*interop.Tag
	for _, t := range tags {
		tagsProto = append(tagsProto, &interop.Tag{Key: t.Key.Name(), Value: t.Value})
	}
	return tagsProto
}
