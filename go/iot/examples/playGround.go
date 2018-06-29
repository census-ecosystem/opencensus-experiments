package main

import (
	"fmt"
	"go.opencensus.io/stats"
	"log"
)

type Vertex struct {
	Lat, Long float64
}

var m map[string]interface{}
var (
	soundMeasure = stats.Int64(name)
)

func main() {
	m = make(map[string]interface{})
	test := make(map[string]interface{})
	test["T"] = "String"
	test["A"] = 1
	m["Bell Labs"] = test
	v, ok := m["Bell Labs"].(map[string]interface{})
	if ok != true {
		log.Fatal(ok)
	}
	var a string
	fmt.Println(v["T"])
	fmt.Println(v["A"])
	fmt.Println(a)
}
