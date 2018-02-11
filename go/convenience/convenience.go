// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package convenience

import (
	"fmt"
	"go.opencensus.io/stats"
	"log"
	"time"
)

type Int64Recorder func(int64) stats.Measurement

func NewCounter(prefix, name, desc string) (Int64Recorder, *stats.View) {
	fullname := fmt.Sprintf("%s/%s", prefix, name)
	m, err := stats.NewMeasureInt64(fullname, desc, "")
	if err != nil {
		log.Panic("unable to create measure", err)
	}
	v, err := stats.NewView(fullname, desc, nil, m, stats.SumAggregation{}, stats.Cumulative{})
	if err != nil {
		log.Panic("unable to create view", err)
	}
	return m.M, v
}

func NewGauge(prefix, name, desc string) (Int64Recorder, *stats.View) {
	fullname := fmt.Sprintf("%s/%s", prefix, name)
	m, err := stats.NewMeasureInt64(fullname, desc, "")
	if err != nil {
		log.Panic("unable to create measure", err)
	}
	// TODO: change aggregation function to Latest
	v, err := stats.NewView(fullname, desc, nil, m, stats.SumAggregation{}, stats.Cumulative{})
	if err != nil {
		log.Panic("unable to create view", err)
	}
	return m.M, v
}

type Stopwatch struct {
	m            *stats.MeasureFloat64
	Distribution *stats.View
}

func NewTimer(prefix, desc string) Stopwatch {
	m, err := stats.NewMeasureFloat64(fmt.Sprintf("%s/%s", prefix, "time"), desc, "")
	if err != nil {
		log.Fatal("unable to create measure", err)
	}
	v, err := stats.NewView(fmt.Sprintf("%s/%s", prefix, "time"), desc, nil, m, defaultTimeDistribution(), stats.Cumulative{})
	if err != nil {
		log.Fatal("unable to create view", err)
	}
	return Stopwatch{m: m, Distribution: v}
}

type Stopper func() stats.Measurement

func (sw Stopwatch) Start() Stopper {
	start := time.Now()
	return func() stats.Measurement {
		end := time.Now()
		return sw.m.M(float64(end.Sub(start)) / float64(time.Microsecond))
	}
}

func defaultTimeDistribution() stats.DistributionAggregation {
	return stats.DistributionAggregation{0.0, 0.5, 1.0, 0.5e1, 1e1, 0.5e2, 1e2, 0.5e3, 1e3, 1.5e3, 1e4, 1.5e4, 1e5, 1.5e5, 1e6, 1e7, 1e8}
}
