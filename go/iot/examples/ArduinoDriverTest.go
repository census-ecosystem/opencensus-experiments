package main

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
)

func main() {
	test := &protocol.Test {
		Label: proto.String("hello"),
		Type:  proto.Int32(17),
		Reps:  []int64{1, 2, 3},
	}
	data, err := proto.Marshal(test)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	newTest := &protocol.Test{}
	err = proto.Unmarshal(data, newTest)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	// Now test and newTest contain the same data.
	if test.GetLabel() != newTest.GetLabel() {
		log.Fatalf("data mismatch %q != %q", test.GetLabel(), newTest.GetLabel())
	} else{
		log.Print("Success!\n")
	}
	// etc.
}