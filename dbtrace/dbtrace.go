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

package dbtrace

import (
	"context"
	"go.opencensus.io/trace"
	"go.opencensus.io/stats"
	"sync/atomic"
	"database/sql"
	"github.com/ramonza/opencensus-experiments/convenience"
)

const (
	queryOperation = "opencensus.io/db/query"
	execOperation  = "opencensus.io/db/exec"
)

var (
	RowsPerQuery       = convenience.NewCounter(queryOperation, "rows", "Number of rows per query")
	QueryTime          = convenience.NewTimer(queryOperation, "Time spent reading and processing query results, in microseconds")
	ActiveQueries      = convenience.NewGauge(queryOperation, "active", "Queries active")
	activeQueriesCount int32

	ExecTime         = convenience.NewTimer(execOperation, "Time spent reading and processing query results, in microseconds")
	ActiveExecs      = convenience.NewGauge(execOperation, "active", "Queries active")
	RowsAffected     = convenience.NewCounter(execOperation, "rows", "Rows affected")
	activeExecsCount int32
)

type Query struct {
	Span     *trace.Span
	Err      error
	Rows     *sql.Rows
	Query    string
	rowsRead int32
	sr       convenience.StopwatchRun
}

func StartQuery(ctx context.Context, query string) *Query {
	ctx = trace.StartSpan(ctx, queryOperation)
	span := trace.FromContext(ctx)
	span.SetAttributes(trace.StringAttribute{Key: "query", Value: query})
	span.SetStackTrace()
	stats.Record(ctx, ActiveQueries.M(int64(atomic.AddInt32(&activeQueriesCount, 1))))
	return &Query{sr: QueryTime.Start(), Span: span, Query: query}
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
		q.Span.Print("Next result set")
		return true
	}
	return false
}

func (q *Query) End(ctx context.Context) {
	atomic.AddInt32(&activeQueriesCount, -1)
	q.Span.SetStatus(statusFromError(q.Err))
	q.Span.End()
	stats.Record(ctx,
		RowsPerQuery.M(int64(q.rowsRead)),
		q.sr.Stop(),
		ActiveQueries.M(int64(atomic.AddInt32(&activeQueriesCount, -1))))
}

type Exec struct {
	Span   *trace.Span
	Query  string
	Result sql.Result
	Err    error
	sr     convenience.StopwatchRun
}

func StartExec(ctx context.Context, stmt string) *Exec {
	ctx = trace.StartSpan(ctx, execOperation)
	span := trace.FromContext(ctx)
	span.SetAttributes(trace.StringAttribute{Key: "query", Value: stmt})
	span.SetStackTrace()
	stats.Record(ctx, ActiveExecs.M(int64(atomic.AddInt32(&activeExecsCount, 1))))
	return &Exec{sr: ExecTime.Start(), Query: stmt, Span: span}
}

func (e *Exec) End(ctx context.Context) {
	atomic.AddInt32(&activeExecsCount, -1)
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
		e.sr.Stop(),
		RowsAffected.M(rowsAffected),
		ActiveExecs.M(int64(atomic.AddInt32(&activeExecsCount, -1))))
}

func statusFromError(err error) trace.Status {
	if err != nil {
		return trace.Status{Code: 2, Message: err.Error()}
	} else {
		return trace.Status{Code: 0}
	}
}
