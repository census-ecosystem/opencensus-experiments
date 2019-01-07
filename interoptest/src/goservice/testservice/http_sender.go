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
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace/propagation"
	"golang.org/x/net/context"
	"goservice/genproto"
	"io/ioutil"
	"net/http"
)

// HttpSender is the type that handles test requests over HTTP.
type HttpSender struct {
	b3Client *http.Client
	tcClient *http.Client
}

func newHttpClient(propagation propagation.HTTPFormat) *http.Client {
	return &http.Client{Transport: &ochttp.Transport{Propagation: propagation}}
}

func newB3Client() *http.Client {
	return newHttpClient(&b3.HTTPFormat{})
}

func newTcClient() *http.Client {
	return newHttpClient(&tracecontext.HTTPFormat{})
}

// NewHttpSender just creates HTTP Clients that sends Test request over HTTP
// It creates two clients, one for B3 propagation and other for TraceContext propagation.
func NewHttpSender() *HttpSender {
	return &HttpSender{b3Client: newB3Client(), tcClient: newTcClient()}
}

// Send sends http request to a server specified by serviceHop.
func (hs *HttpSender) Send(ctx context.Context, serviceHop interop.ServiceHop, req *interop.TestRequest) (*interop.TestResponse, error) {
	res := interop.TestResponse{
		Id: req.GetId(),
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	url := fmt.Sprintf("http://%s:%d/test/request", serviceHop.Service.Host, serviceHop.Service.Port)

	if serviceHop.GetService().GetSpec().GetPropagation() == interop.Spec_B3_FORMAT_PROPAGATION {
		resp, err = hs.b3Client.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	} else if serviceHop.GetService().GetSpec().GetPropagation() == interop.Spec_TRACE_CONTEXT_FORMAT_PROPAGATION {
		resp, err = hs.tcClient.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	}
	if resp != nil {
		defer resp.Body.Close()
		respData, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = proto.Unmarshal(respData, &res)
			if err == nil {
				return &res, nil
			}
		}
	}
	return nil, err
}
