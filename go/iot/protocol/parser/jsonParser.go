package parser

import (
	"encoding/json"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
)

type JsonParser struct {
}

func (parser *JsonParser) DecodeMeasurement(input []byte) (protocol.MeasureArgument, error) {
	var output protocol.MeasureArgument
	decodeErr := json.Unmarshal(input, &output)
	return output, decodeErr
}

func (parser *JsonParser) EncodeResponse(myResponse *protocol.Response) ([]byte, error) {
	b, err := json.Marshal(myResponse)
	return b, err
}
