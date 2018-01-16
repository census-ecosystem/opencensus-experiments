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

package dbtrace_test

import (
	"github.com/ramonza/opencensus-experiments/dbtrace"
	"context"
	"database/sql"
	"log"
	"github.com/ramonza/opencensus-experiments/testing/mocktrace"
	"github.com/ramonza/opencensus-experiments/testing/mockstats"
	"time"
	"go.opencensus.io/stats"
)

func init() {
	mocktrace.RegisterExporter()
	mockstats.RegisterExporter()
}

func ExampleDbtrace() {
	dbtrace.ExecTime.Distribution.Subscribe()
	dbtrace.QueryTime.Distribution.Subscribe()

	ctx := context.Background()

	db, err := sql.Open("", "")
	if err != nil {
		log.Fatal(err)
	}

	ps, _ := db.PrepareContext(ctx, "")

	_ = ps

	exec := dbtrace.StartExec(ctx, "CREATE TABLE")
	exec.Result, exec.Err = db.ExecContext(ctx, exec.Query)
	exec.End(ctx)

	spans := mocktrace.Spans("opencensus.io/db/exec")
	if len(spans) != 1 {
		log.Fatalf("expected a single span: %#v", spans)
	}
	span := spans[0]
	if span.Attributes["query"] != "CREATE TABLE" {
		log.Fatalf("expected query attribute: %#v", span)
	}

	q := dbtrace.StartQuery(ctx, "")
	q.Rows, q.Err = db.QueryContext(ctx, q.Query)


	if q.Err != nil {
		for q.NextRow() {
			q.Rows.Scan()
		}
	}

	q.End(ctx)

	stats.SetReportingPeriod(100 * time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	e := <-mockstats.Exported(dbtrace.ExecTime.Distribution)
	if e.View == dbtrace.ExecTime.Distribution {

	}
}
