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

	spanID10 = trace.SpanID{0, 0, 0, 0, 0, 0, 0, 0}
	spanID11 = trace.SpanID{1, 1, 1, 1, 1, 1, 1, 1}
	spanID12 = trace.SpanID{2, 2, 2, 2, 2, 2, 2, 2}
	spanID13 = trace.SpanID{3, 3, 3, 3, 3, 3, 3, 3}
	spanID14 = trace.SpanID{4, 4, 4, 4, 4, 4, 4, 4}
	spanID21 = trace.SpanID{5, 5, 5, 5, 5, 5, 5, 5}
	spanID22 = trace.SpanID{6, 6, 6, 6, 6, 6, 6, 6}
	spanID15 = trace.SpanID{7, 7, 7, 7, 7, 7, 7, 7}
	spanID16 = trace.SpanID{8, 8, 8, 8, 8, 8, 8, 8}

	// trace 1
	span10 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID11[:], ParentSpanId: spanID10[:]} // root with 0 spanId
	span11 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID11[:]}                            // root nil spanId
	span12 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID12[:], ParentSpanId: spanID11[:]} // child
	span13 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID13[:], ParentSpanId: spanID12[:]} // grandchild 1
	span14 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID14[:], ParentSpanId: spanID12[:]} // grandchild 2
	span15 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID15[:], ParentSpanId: spanID14[:]} // greatgrandchild 1
	span16 = &tracepb.Span{TraceId: traceID1[:], SpanId: spanID16[:], ParentSpanId: spanID15[:]} // greatgreatgrandchild 1

	// trace 2
	span21 = &tracepb.Span{TraceId: traceID2[:], SpanId: spanID21[:]}                            // root
	span22 = &tracepb.Span{TraceId: traceID2[:], SpanId: spanID22[:], ParentSpanId: spanID21[:]} // child
	span23 = &tracepb.Span{TraceId: traceID2[:], SpanId: spanID22[:]}                            // another root
)

func TestReconstructTraces(t *testing.T) {
	want := map[trace.TraceID]*SimpleSpan{
		traceID1: {
			traceID: traceID1,
			spanID:  spanID11,
			children: map[trace.SpanID]*SimpleSpan{
				spanID12: {
					traceID: traceID1,
					spanID:  spanID12,
					children: map[trace.SpanID]*SimpleSpan{
						spanID13: {
							traceID:  traceID1,
							spanID:   spanID13,
							children: map[trace.SpanID]*SimpleSpan{},
						},
						spanID14: {
							traceID: traceID1,
							spanID:  spanID14,
							children: map[trace.SpanID]*SimpleSpan{
								spanID15: {
									traceID: traceID1,
									spanID:  spanID15,
									children: map[trace.SpanID]*SimpleSpan{
										spanID16: {
											traceID:  traceID1,
											spanID:   spanID16,
											children: map[trace.SpanID]*SimpleSpan{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		traceID2: {
			traceID:  traceID2,
			spanID:   spanID21,
			children: map[trace.SpanID]*SimpleSpan{},
		},
	}

	spanCombinations := [][]*tracepb.Span{
		[]*tracepb.Span{span10, span12, span13, span14, span15, span16, span21},
		[]*tracepb.Span{span12, span13, span14, span15, span16, span10, span21},
		[]*tracepb.Span{span13, span14, span15, span16, span12, span10, span21},
		[]*tracepb.Span{span13, span14, span15, span16, span10, span12, span21},
		[]*tracepb.Span{span10, span13, span14, span15, span16, span12, span21},
		[]*tracepb.Span{span15, span16, span11, span12, span13, span14, span21},
		[]*tracepb.Span{span12, span13, span14, span11, span15, span16, span21},
		[]*tracepb.Span{span13, span14, span12, span15, span16, span11, span21},
		[]*tracepb.Span{span13, span14, span11, span15, span16, span12, span21},
		[]*tracepb.Span{span15, span16, span11, span13, span14, span12, span21},
	}

	for i, combination := range spanCombinations {
		got, errs := ReconstructTraces(combination...)
		if len(errs) > 0 {
			t.Fatalf("Failed to reconstruct trace: %v", errs)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Reconstruct trace\n\tGot  %+v\n\tWant %+v, combination %d", got, want, i)
		}
	}
}

func TestReconstructTraces_alreadyExists(t *testing.T) {
	roots, errs := ReconstructTraces(span11, span11)
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
	roots, errs := ReconstructTraces(span11, span12, span13, span14, span22)
	wantErrs := map[trace.TraceID]error{traceID2: errOrphanSpan}
	if !reflect.DeepEqual(errs, wantErrs) {
		t.Fatalf("Want error\n\tGot  %+v\n\tWant %+v", errs, wantErrs)
	}
	wantRoots := map[trace.TraceID]*SimpleSpan{
		traceID1: {
			traceID: traceID1,
			spanID:  spanID11,
			children: map[trace.SpanID]*SimpleSpan{
				spanID12: {
					traceID: traceID1,
					spanID:  spanID12,
					children: map[trace.SpanID]*SimpleSpan{
						spanID13: {
							traceID:  traceID1,
							spanID:   spanID13,
							children: map[trace.SpanID]*SimpleSpan{},
						},
						spanID14: {
							traceID:  traceID1,
							spanID:   spanID14,
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
	roots, errs := ReconstructTraces(span21, span23)
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
	roots, errs := ReconstructTraces(span21, span23, span12)
	wantErrs := map[trace.TraceID]error{traceID1: errOrphanSpan, traceID2: errDuplicatedRootSpan}
	if !reflect.DeepEqual(errs, wantErrs) {
		t.Fatalf("Want error\n\tGot  %+v\n\tWant %+v", errs, wantErrs)
	}
	wantRoots := map[trace.TraceID]*SimpleSpan{}
	if !reflect.DeepEqual(roots, wantRoots) {
		t.Fatalf("Want traces\n\tGot  %+v\n\tWant %+v", roots, wantRoots)
	}
}

func TestGroupByTraceId(t *testing.T) {
	spanCombinations := [][]*tracepb.Span{
		[]*tracepb.Span{span11, span12, span13, span14, span15, span16, span21},
		[]*tracepb.Span{span12, span13, span14, span15, span16, span11, span21},
		[]*tracepb.Span{span13, span14, span15, span16, span12, span11, span21},
		[]*tracepb.Span{span13, span14, span15, span16, span11, span12, span21},
		[]*tracepb.Span{span11, span13, span14, span15, span16, span12, span21},
		[]*tracepb.Span{span15, span16, span11, span12, span13, span14, span21},
		[]*tracepb.Span{span12, span13, span14, span11, span15, span16, span21},
		[]*tracepb.Span{span13, span14, span12, span15, span16, span11, span21},
		[]*tracepb.Span{span13, span14, span11, span15, span16, span12, span21},
		[]*tracepb.Span{span15, span16, span11, span13, span14, span12, span21},
	}

	want := map[trace.TraceID][]*tracepb.Span{
		traceID1: {span11, span12, span13, span14, span15, span16},
		traceID2: {span21},
	}
	for i, combination := range spanCombinations {
		got := groupSpansByTraceID(combination)
	outer:
		for tId, spans := range want {
			if got[tId] == nil {
				t.Fatalf("grouping span by tracdId failed for Id:%s, combination: %d", tId.String(), i)
			}
			for _, wantSpan := range spans {
				for _, gotSpan := range got[tId] {
					if wantSpan == gotSpan {
						continue outer
					}
				}
				t.Fatalf("grouping span by tracdId failed, span %v not found, combination: %d", wantSpan, i)
			}
		}
	}
}
