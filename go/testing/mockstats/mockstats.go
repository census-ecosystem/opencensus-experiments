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

// TODO: High-level file comment.
package mockstats

import (
	"go.opencensus.io/stats"
	"log"
	"sync"
)

type Exporter struct {
	m *sync.Map
}

var e Exporter

func RegisterExporter() {
	stats.RegisterExporter(&e)
}

func (e Exporter) Export(viewData *stats.ViewData) {
	select {
	case e.ch(viewData.View) <- viewData:
		break
	default:
		log.Fatalf("mockstats.Exporter buffer full")
	}
}
func (e Exporter) ch(v *stats.View) chan *stats.ViewData {
	val, ok := e.m.Load(v)
	if !ok {
		val, _ = e.m.LoadOrStore(v, make(chan *stats.ViewData, 1000))
	}
	ch := val.(chan *stats.ViewData)
	return ch
}

func Exported(v *stats.View) <-chan *stats.ViewData {
	return e.ch(v)
}
