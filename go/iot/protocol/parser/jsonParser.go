package parser

import (
	"encoding/json"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
)

type JsonParser struct {
}

func (parser *JsonParser) Parse(input []byte) (protocol.MeasureArgument, error) {
	var output protocol.MeasureArgument
	decodeErr := json.Unmarshal(input, &output)
	return output, decodeErr
}
