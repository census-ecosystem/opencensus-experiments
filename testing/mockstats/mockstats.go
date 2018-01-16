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
