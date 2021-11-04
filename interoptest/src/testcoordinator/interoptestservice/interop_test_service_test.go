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

package interoptestservice_test

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator/interoptestservice"
)

func TestRequestsOverHTTP(t *testing.T) {
	h := new(interoptestservice.ServiceImpl)
	tests := []struct {
		reqWire     string // The request's wire data
		wantResWire string // The response's wire format
	}{

		// OPTIONS request
		{
			reqWire: `OPTIONS / HTTP/1.1
Host: *

`,
			wantResWire: "HTTP/1.1 200 OK\r\n" +
				"Connection: close\r\n" +
				"Access-Control-Allow-Headers: *\r\n" +
				"Access-Control-Allow-Methods: *\r\n" +
				"Access-Control-Allow-Origin: *\r\n\r\n",
		},

		// GET: bad path
		{
			reqWire: `GET / HTTP/1.1
Host: foo
Content-Length: 0

`,
			wantResWire: "HTTP/1.1 400 Bad Request\r\n" +
				"Connection: close\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"X-Content-Type-Options: nosniff\r\n\r\n" +
				"Expected path of the form: /result/:id\n",
		},

		// GET: good path no id
		{
			reqWire: `GET /result HTTP/1.1

`,
			wantResWire: "HTTP/1.1 400 Bad Request\r\n" +
				"Connection: close\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"X-Content-Type-Options: nosniff\r\n\r\n" +
				"Expected path of the form: /result/:id\n",
		},

		// GET: good path with proper id
		{
			reqWire: `GET /result/1 HTTP/1.1

`,
			wantResWire: "HTTP/1.1 200 OK\r\n" +
				"Connection: close\r\n" +
				"Content-Type: application/json\r\n\r\n" +
				`{"id":1,"status":{}}`,
		},
		// POST: no body
		{
			reqWire: `POST /result HTTP/1.1

`,
			wantResWire: "HTTP/1.1 400 Bad Request\r\n" +
				"Connection: close\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"X-Content-Type-Options: nosniff\r\n\r\n" +
				"Failed to JSON unmarshal interop.InteropResultRequest: unexpected end of JSON input\n",
		},

		// POST: body with content length to accepted route
		{
			reqWire: `POST /result HTTP/1.1
Content-Length: 9
Content-Type: application/json

{"id":10}
`,
			wantResWire: "HTTP/1.1 200 OK\r\n" +
				"Connection: close\r\n" +
				"Content-Type: application/json\r\n\r\n" +
				`{"id":10,"status":{}}`,
		},

		// POST: body with no content length
		{
			// Using a string concatenation here because for "streamed"/"chunked"
			// requests, we have to ensure that the last 2 bytes before EOF are
			// strictly "\r\n" lest a "malformed chunked encoding" error.
			reqWire: "POST /result HTTP/1.1\r\n" +
				"Host: golang.org\r\n" +
				"Content-Type: application/json\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"Accept-Encoding: gzip\r\n\r\n" +
				"b\r\n" +
				"{\"id\":8888}\r\n" +
				"0\r\n\r\n",
			wantResWire: "HTTP/1.1 200 OK\r\n" +
				"Connection: close\r\n" +
				"Content-Type: application/json\r\n\r\n" +
				`{"id":8888,"status":{}}`,
		},

		// POST: body with content length to non-existent route
		{
			reqWire: `POST /results HTTP/1.1
Content-Length: 9
Content-Type: application/json

{"id":10}
`,
			wantResWire: "HTTP/1.1 404 Not Found\r\n" +
				"Connection: close\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
                                "X-Content-Type-Options: nosniff\r\n\r\n" +
                                "Unmatched route: /results\n" +
                                "Only accepting /result and /run\n",
		},
	}

	for i, tt := range tests {
		req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(tt.reqWire)))
		if err != nil {
			t.Errorf("#%d unexpected error parsing request: %v", i, err)
			continue
		}

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		gotResBlob, _ := httputil.DumpResponse(rec.Result(), true)
		gotRes := string(gotResBlob)
		if gotRes != tt.wantResWire {
			t.Errorf("#%d non-matching responses\nGot:\n%q\nWant:\n%q", i, gotRes, tt.wantResWire)
		}
	}
}
