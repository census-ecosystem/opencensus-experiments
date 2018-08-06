package parser

import (
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
)

type Parser interface {
	Decode(input []byte) (protocol.MeasureArgument, error)
	Encode(myResponse *protocol.Response) ([]byte, error)
}
