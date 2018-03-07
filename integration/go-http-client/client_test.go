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

package interop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"golang.org/x/net/context"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/google"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	pb "github.com/census-instrumentation/opencensus-experiments/integration/proto"
)

var setups = []struct {
	name         string
	envAddrKey   string
	fallbackAddr string
}{
	{name: "GoClient-GoServer-HTTP", envAddrKey: "GO_HTTP_SERVER_ADDR", fallbackAddr: ":9900"},
	// {name: "GoClient-JavaServer-HTTP", envName: "JAVA_SERVER_ADDR", fallbackAddr: ":9901"},
	// {name: "GoClient-PythonServer-HTTP", envName: "PYTHON_SERVER_ADDR", fallbackAddr: ":9902"},
}

var propagations = []string{"b3", "google", "tracecontext"}

func TestInterop(t *testing.T) {
	for _, setup := range setups {
		addr := os.Getenv(setup.envAddrKey)
		if addr == "" {
			addr = setup.fallbackAddr
		}

		for _, propagation := range propagations {
			t.Run(setup.name+"/propagation="+propagation, func(tt *testing.T) {
				runInteropTest(tt, addr, propagation)
			})
		}
	}
}

func runInteropTest(t *testing.T, host, propagationStr string) {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/?p=%s", host, propagationStr), nil)
	if err != nil {
		t.Fatalf("go-HTTP client HTTP test err: %v", err)
	}

	// 1. Create some tags
	ctx, err := tag.New(context.Background(),
		tag.Insert(mustKey("operation"), "interop-test"),
		tag.Insert(mustKey("project"), "open-census"),
	)
	if err != nil {
		t.Fatalf("tag.New err: %v", err)
	}
	req = req.WithContext(ctx)

	// 2. Create a span with a traceID, spanID
	inSpanCtx := trace.SpanContext{
		TraceID:      trace.TraceID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		SpanID:       trace.SpanID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F},
		TraceOptions: 1,
	}

	var hf propagation.HTTPFormat
	switch propagationStr {
	case "b3":
		hf = new(b3.HTTPFormat)
	case "google":
		hf = new(google.HTTPFormat)
	case "tracecontext":
		hf = new(tracecontext.HTTPFormat)
	}

	hf.SpanContextToRequest(inSpanCtx, req)

	// t.Logf("%q: %+v\n", propagationStr, req)
	httpClient := &http.Client{
		Transport: &ochttp.Transport{
			Sampler:     trace.AlwaysSample(),
			Propagation: hf,
		},
	}

	res, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("httpClient.Do err: %v", err)
	}
	if code := res.StatusCode; code < 200 || code > 299 {
		t.Fatalf("httpClient StatusCode(%d) != 2XX status: %q", code, res.Status)
	}
	blob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Response.Body read err: %v", err)
	}
	eres := new(pb.EchoResponse)
	if err := json.Unmarshal(blob, eres); err != nil {
		t.Fatalf("UnmarshalJSON err: %v", err)
	}

	if gti, wti := eres.TraceId, inSpanCtx.TraceID[:]; !bytes.Equal(gti, wti) {
		t.Errorf("TraceID:\ngot= (% X)\nwant=(% X)", gti, wti)
	}
	if gsi, wsi := eres.SpanId, inSpanCtx.SpanID[:]; !bytes.Equal(gsi, wsi) {
		t.Errorf("SpanID:\ngot= (% X)\nwant=(% X)", gsi, wsi)
	}
}

func mustKey(key string) tag.Key {
	k, err := tag.NewKey(key)
	if err != nil {
		log.Fatalf("tag.NewKey: %q err: %v", key, err)
	}
	return k
}
