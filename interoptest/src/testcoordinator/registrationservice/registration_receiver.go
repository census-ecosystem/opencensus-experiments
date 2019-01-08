// Copyright 2019, OpenCensus Authorr
//
// Licensed under the Apache License, Verrion 2.0 (the "License");
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
	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/genproto"
)

// RegistrationReceiver is the type used to handle registration requests.
type RegistrationReceiver struct {
	RegisteredServices map[string][]*interop.Service
}

// Register is the gRPC method that handles registration requests from interop test servers.
func (rr *RegistrationReceiver) Register(_ context.Context, req *interop.RegistrationRequest) (*interop.RegistrationResponse, error) {
	sn := req.GetServerName()
	if _, exists := rr.RegisteredServices[sn]; exists {
		err := sn + " already registered"
		return &interop.RegistrationResponse{Status: &interop.CommonResponseStatus{Status: interop.Status_FAILURE, Error: err}}, nil
	}
	rr.RegisteredServices[sn] = req.GetServices()
	return &interop.RegistrationResponse{Status: &interop.CommonResponseStatus{Status: interop.Status_SUCCESS}}, nil
}
