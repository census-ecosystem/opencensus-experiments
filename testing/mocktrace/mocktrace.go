package mocktrace

import "go.opencensus.io/trace"

type Exporter []*trace.SpanData

var e Exporter

func RegisterExporter() {
	trace.RegisterExporter(&e)
}

func Spans(name string) []*trace.SpanData {
	if name == "" {
		return e[:]
	}
	spans := make([]*trace.SpanData, 0)
	for _, span := range e {
		if span.Name == name {
			spans = append(spans, span)
		}
	}
	return spans
}

func (e *Exporter) Export(s *trace.SpanData) {
	*e = append(*e, s)
}
