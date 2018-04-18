package store

import (
	"time"
	"sync"
	"github.com/census-instrumentation/opencensus-experiments/go/tracelyzer/tracelyzerpb"
	"log"
)

type Trace struct {
	TraceID    string
	Spans      []*tracelyzerpb.Span
	expireTime time.Time
}

func (t *Trace) Root() *tracelyzerpb.Span {
	for _, sp := range t.Spans {
		if len(sp.ParentSpanId) == 0 {
			return sp
		}
	}
	log.Printf("No root: %x", t.TraceID)
	return nil
}

type Store struct {
	mu         sync.Mutex
	expiryTime time.Duration
	pending    map[string]*Trace
	traces     chan<- *Trace
}

func NewStore(expiryTime time.Duration, traces chan<- *Trace) *Store {
	s := &Store{
		pending:    make(map[string]*Trace),
		traces:     traces,
		expiryTime: expiryTime,
	}
	go s.expireLoop()
	return s
}

func (s *Store) PutSpan(traceID string, span *tracelyzerpb.Span) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireAt := time.Now().Add(s.expiryTime)
	if len(span.ParentSpanId) == 0 {
		// expect root spans to be sent last, so wait less time
		expireAt = time.Now().Add(100 * time.Millisecond)
	}

	trace, ok := s.pending[traceID]
	if !ok {
		trace = &Trace{
			TraceID:    traceID,
			expireTime: expireAt,
		}
		s.pending[traceID] = trace
	}
	trace.Spans = append(trace.Spans, span)

	if trace.expireTime.Before(expireAt) {
		trace.expireTime = expireAt
	}
}

func (s *Store) expireLoop() {
	for {
		<-time.After(time.Second)
		s.mu.Lock()
		t := time.Now()
		var expired []*Trace
		for k, v := range s.pending {
			if v.expireTime.Before(t) {
				expired = append(expired, v)
				delete(s.pending, k)
			}
		}
		s.mu.Unlock()
		for _, trace := range expired {
			s.traces <- trace
		}
	}
}
