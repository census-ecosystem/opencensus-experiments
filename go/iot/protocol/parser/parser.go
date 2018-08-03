package parser

import (
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
)

type Parser interface {
	Parse(input []byte) (protocol.MeasureArgument, error)
}
