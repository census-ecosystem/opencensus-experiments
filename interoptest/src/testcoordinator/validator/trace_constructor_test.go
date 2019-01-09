package validator

import (
	"go.opencensus.io/trace"
	"reflect"
	"testing"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
)

var (
	traceID1 = trace.TraceID{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	traceID2 = trace.TraceID{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}

	spanID1 = trace.SpanID{1, 1, 1, 1, 1, 1, 1, 1}
	spanID2 = trace.SpanID{2, 2, 2, 2, 2, 2, 2, 2}
	spanID3 = trace.SpanID{3, 3, 3, 3, 3, 3, 3, 3}
	spanID4 = trace.SpanID{4, 4, 4, 4, 4, 4, 4, 4}
	spanID5 = trace.SpanID{5, 5, 5, 5, 5, 5, 5, 5}
	spanID6 = trace.SpanID{6, 6, 6, 6, 6, 6, 6, 6}

	// trace 1
	span1 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID1[:]}                           // root
	span2 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID2[:], ParentSpanId: spanID1[:]} // child
	span3 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID3[:], ParentSpanId: spanID2[:]} // grandchild 1
	span4 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID4[:], ParentSpanId: spanID2[:]} // grandchild 2

	// trace 2
	span5 = &tracepb.Span{TraceId: traceID2[:], SpanId: spanID5[:]}                           // root
	span6 = &tracepb.Span{TraceId: traceID2[:], SpanId: spanID6[:], ParentSpanId: spanID5[:]} // child
	span7 = &tracepb.Span{TraceId: traceID2[:], SpanId: spanID6[:]}                           // another root
)

func TestReconstructTraces(t *testing.T) {
	got, errs := ReconstructTraces(span3, span4, span1, span5, span2)
	if len(errs) > 0 {
		t.Fatalf("Failed to reconstruct trace: %v", errs)
	}
	want := map[trace.TraceID]*SimpleSpan{
		traceID1: {
			traceID: traceID1,
			spanID:  spanID1,
			children: map[trace.SpanID]*SimpleSpan{
				spanID2: {
					traceID: traceID1,
					spanID:  spanID2,
					children: map[trace.SpanID]*SimpleSpan{
						spanID3: {
							traceID:  traceID1,
							spanID:   spanID3,
							children: map[trace.SpanID]*SimpleSpan{},
						},
						spanID4: {
							traceID:  traceID1,
							spanID:   spanID4,
							children: map[trace.SpanID]*SimpleSpan{},
						},
					},
				},
			},
		},
		traceID2: {
			traceID:  traceID2,
			spanID:   spanID5,
			children: map[trace.SpanID]*SimpleSpan{},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Reconstruct trace\n\tGot  %+v\n\tWant %+v", got, want)
	}
}

func TestReconstructTraces_alreadyExists(t *testing.T) {
	roots, errs := ReconstructTraces(span1, span1)
	wantErrs := map[trace.TraceID]error{traceID1: errAlreadyExists}
	if !reflect.DeepEqual(errs, wantErrs) {
		t.Fatalf("Want error\n\tGot  %+v\n\tWant %+v", errs, wantErrs)
	}
	wantRoots := map[trace.TraceID]*SimpleSpan{}
	if !reflect.DeepEqual(roots, wantRoots) {
		t.Fatalf("Want traces\n\tGot  %+v\n\tWant %+v", roots, wantRoots)
	}
}

func TestReconstructTraces_orphan(t *testing.T) {
	roots, errs := ReconstructTraces(span1, span2, span3, span4, span6)
	wantErrs := map[trace.TraceID]error{traceID2: errOrphanSpan}
	if !reflect.DeepEqual(errs, wantErrs) {
		t.Fatalf("Want error\n\tGot  %+v\n\tWant %+v", errs, wantErrs)
	}
	wantRoots := map[trace.TraceID]*SimpleSpan{
		traceID1: {
			traceID: traceID1,
			spanID:  spanID1,
			children: map[trace.SpanID]*SimpleSpan{
				spanID2: {
					traceID: traceID1,
					spanID:  spanID2,
					children: map[trace.SpanID]*SimpleSpan{
						spanID3: {
							traceID:  traceID1,
							spanID:   spanID3,
							children: map[trace.SpanID]*SimpleSpan{},
						},
						spanID4: {
							traceID:  traceID1,
							spanID:   spanID4,
							children: map[trace.SpanID]*SimpleSpan{},
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(roots, wantRoots) {
		t.Fatalf("Want traces\n\tGot  %+v\n\tWant %+v", roots, wantRoots)
	}
}

func TestReconstructTraces_duplicatedRoots(t *testing.T) {
	roots, errs := ReconstructTraces(span5, span7)
	wantErrs := map[trace.TraceID]error{traceID2: errDuplicatedRootSpan}
	if !reflect.DeepEqual(errs, wantErrs) {
		t.Fatalf("Want error\n\tGot  %+v\n\tWant %+v", errs, wantErrs)
	}
	wantRoots := map[trace.TraceID]*SimpleSpan{}
	if !reflect.DeepEqual(roots, wantRoots) {
		t.Fatalf("Want traces\n\tGot  %+v\n\tWant %+v", roots, wantRoots)
	}
}

func TestReconstructTraces_mixedErrs(t *testing.T) {
	roots, errs := ReconstructTraces(span5, span7, span2)
	wantErrs := map[trace.TraceID]error{traceID1: errOrphanSpan, traceID2: errDuplicatedRootSpan}
	if !reflect.DeepEqual(errs, wantErrs) {
		t.Fatalf("Want error\n\tGot  %+v\n\tWant %+v", errs, wantErrs)
	}
	wantRoots := map[trace.TraceID]*SimpleSpan{}
	if !reflect.DeepEqual(roots, wantRoots) {
		t.Fatalf("Want traces\n\tGot  %+v\n\tWant %+v", roots, wantRoots)
	}
}
