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

// TODO: High-level file comment.

package convenience

import (
	"fmt"
	"go.opencensus.io/stats"
	"log"
	"time"
)

type Counter struct {
	*stats.MeasureInt64
	Total *stats.View
}

func NewCounter(prefix, name, desc string) Counter {
	fullname := fmt.Sprintf("%s/%s", prefix, name)
	m, err := stats.NewMeasureInt64(fullname, desc, "")
	if err != nil {
		log.Fatal("unable to create measure", err)
	}
	v, err := stats.NewView(fullname, desc, nil, m, stats.SumAggregation{}, stats.Cumulative{})
	if err != nil {
		log.Fatal("unable to create view", err)
	}
	return Counter{MeasureInt64: m, Total: v}
}

type Guage struct {
	*stats.MeasureInt64
	Mean *stats.View
}

func NewGauge(prefix, name, desc string) Guage {
	fullname := fmt.Sprintf("%s/%s", prefix, name)
	m, err := stats.NewMeasureInt64(fullname, desc, "")
	if err != nil {
		log.Fatal("unable to create measure", err)
	}
	// TODO: add default view once #318 is fixed
	return Guage{MeasureInt64: m, Mean: nil}
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
	v, err := stats.NewView(fmt.Sprintf("%s/%s", prefix, "time"), desc, nil, m, stats.DistributionAggregation{}, stats.Cumulative{})
	if err != nil {
		log.Fatal("unable to create view", err)
	}
	return Stopwatch{m: m, Distribution: v}
}

type StopwatchRun struct {
	m     *stats.MeasureFloat64
	start time.Time
}

func (sw Stopwatch) Start() StopwatchRun {
	return StopwatchRun{m: sw.m, start: time.Now()}
}

func (sws StopwatchRun) Stop() stats.Measurement {
	end := time.Now()
	return sws.m.M(float64(end.Sub(sws.start)) / float64(time.Microsecond))
}

func defaultTimeDistribution() stats.DistributionAggregation {
	return stats.DistributionAggregation{0.0, 0.5, 1.0, 0.5e1, 1e1, 0.5e2, 1e2, 0.5e3, 1e3, 1.5e3, 1e4, 1.5e4, 1e5, 1.5e5, 1e6, 1e7, 1e8}
}
