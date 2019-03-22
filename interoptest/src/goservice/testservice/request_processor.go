// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testservice

import (
	"fmt"
	"sync"
	"golang.org/x/net/context"
	"goservice/genproto"
	"go.opencensus.io/trace"
)

// RequestProcessor is the type that process test requests.
type RequestProcessor struct {
	httpSender *HttpSender
	grpcSender *GRPCSender
}

var instance *RequestProcessor
var once sync.Once

func (rp *RequestProcessor) getInstance() *RequestProcessor {
	once.Do(func() {
		instance = &RequestProcessor{httpSender: NewHttpSender(), grpcSender: NewGRPCSender()}
	})
	return instance
}

func (rp *RequestProcessor) send(ctx context.Context, serviceHop interop.ServiceHop, req *interop.TestRequest) (*interop.TestResponse, error) {
	if serviceHop.GetService().GetSpec().GetTransport() == interop.Spec_GRPC {
		return rp.grpcSender.Send(ctx, serviceHop, req)
	} else if serviceHop.GetService().GetSpec().GetTransport() == interop.Spec_HTTP {
		return rp.httpSender.Send(ctx, serviceHop, req)
	}
	return nil, invalidTransport
}

func (rp *RequestProcessor) process(ctx context.Context, req *interop.TestRequest) (*interop.TestResponse, error) {
	serviceHops := req.GetServiceHops()
	res := interop.TestResponse{
		Id: req.GetId(),
	}

	span := trace.FromContext(ctx)
	span.AddAttributes(trace.Int64Attribute("reqId", req.GetId()))

	if serviceHops == nil || len(serviceHops) == 0 {
		res.Status = append(res.Status, &interop.CommonResponseStatus{Status: interop.Status_SUCCESS, Error: ""})
	} else {
		// Create a new request.
		newReq := interop.TestRequest{Id: req.Id, Name: req.Name}
		nextServiceHop := req.ServiceHops[0]
		newReq.ServiceHops = req.ServiceHops[1:]
		nextResponse, err := rp.send(ctx, *nextServiceHop, &newReq)
		if err != nil {
			res.Status = append(res.Status, &interop.CommonResponseStatus{Status: interop.Status_FAILURE,
				Error: fmt.Sprintf("failed to send request to nexthop, err:%v", err)})
		} else if nextResponse == nil {
			res.Status = append(res.Status, &interop.CommonResponseStatus{Status: interop.Status_FAILURE,
				Error: "received empty response from nexthop"})
		} else {
			res.Status = append(res.Status, &interop.CommonResponseStatus{Status: interop.Status_SUCCESS, Error: ""})
			res.Status = append(res.Status, nextResponse.Status...)
		}

	}
	return &res, nil
}
