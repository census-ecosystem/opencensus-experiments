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

package dbtrace

import (
	"context"
	"go.opencensus.io/trace"
	"go.opencensus.io/stats"
	"database/sql"
	"github.com/census-instrumentation/opencensus-experiments/go/convenience"
)

const (
	queryOperation = "opencensus.io/db/query"
	execOperation  = "opencensus.io/db/exec"
)

var (
	withRowsPerQuery, RowsPerQuery = convenience.NewCounter(queryOperation, "rows", "Number of rows per query")
	QueryTime                      = convenience.NewTimer(queryOperation, "Time spent reading and processing query results, in microseconds")

	ExecTime                       = convenience.NewTimer(execOperation, "Time spent reading and processing query results, in microseconds")
	withRowsAffected, RowsAffected = convenience.NewCounter(execOperation, "rows", "Rows affected")
)

type Query struct {
	Span     *trace.Span
	Err      error
	Rows     *sql.Rows
	Query    string
	rowsRead int32
	stop     func() stats.Measurement
}

func StartQuery(ctx context.Context, query string) *Query {
	ctx, span := trace.StartSpan(ctx, queryOperation)
	span.SetAttributes(trace.StringAttribute{Key: "query", Value: query})
	return &Query{stop: QueryTime.Start(), Span: span, Query: query}
}

func (q *Query) NextRow() bool {
	if n := q.Rows.Next(); n {
		q.rowsRead++
		return true
	}
	return false
}

func (q *Query) NextResultSet() bool {
	if n := q.Rows.NextResultSet(); n {
		q.Span.Annotate(nil, "Next result set")
		return true
	}
	return false
}

func (q *Query) End(ctx context.Context) {
	q.Span.SetStatus(statusFromError(q.Err))
	q.Span.End()
	stats.Record(ctx,
		withRowsPerQuery(int64(q.rowsRead)),
		q.stop())
}

type Exec struct {
	Span   *trace.Span
	Query  string
	Result sql.Result
	Err    error
	stop   func()stats.Measurement
}

func StartExec(ctx context.Context, stmt string) *Exec {
	ctx, span := trace.StartSpan(ctx, execOperation)
	span.SetAttributes(trace.StringAttribute{Key: "query", Value: stmt})
	return &Exec{stop: ExecTime.Start(), Query: stmt, Span: span}
}

func (e *Exec) End(ctx context.Context) {
	e.Span.SetStatus(statusFromError(e.Err))
	e.Span.End()
	rowsAffected := int64(0)
	if e.Result != nil {
		var err error
		rowsAffected, err = e.Result.RowsAffected()
		if err != nil {
			rowsAffected = 0
		}
	}
	stats.Record(ctx,
		e.stop(),
		withRowsAffected(rowsAffected))
}

func statusFromError(err error) trace.Status {
	if err != nil {
		return trace.Status{Code: 2, Message: err.Error()}
	} else {
		return trace.Status{Code: 0}
	}
}
