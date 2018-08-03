package parser

import (
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
)

type TextParser struct {
}

func (parser *TextParser) Parse(input []byte) (protocol.MeasureArgument, error) {
	var output protocol.MeasureArgument
	// TODO
	return output, nil
}
