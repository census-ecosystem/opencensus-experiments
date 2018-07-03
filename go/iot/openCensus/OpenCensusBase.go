// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openCensus

import (
	"context"
	"log"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/Protocol"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	FLOAT = 0
	INT   = 1
)

type OpenCensusBase struct {
	status      int // Flag to represent whether the system is initialized or not
	ctx         context.Context
	viewNameSet map[string]int           // avoid registering the same view.
	measureMap  map[string]stats.Measure // Store all the measure based on their Name. Used for the future record
	tagKeyMap   map[string][]tag.Key
}

func (census *OpenCensusBase) Initialize(projectId string, reportPeriod int) {
	census.status = 0
	census.ctx = context.Background()
	census.viewNameSet = make(map[string]int)
	census.measureMap = make(map[string]stats.Measure)
	census.tagKeyMap = make(map[string][]tag.Key)
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectId, // Google Cloud Console project ID.
	})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(exporter)
	view.SetReportingPeriod(time.Second * time.Duration(reportPeriod))
}

// Return whether the view has already been registered
func (census *OpenCensusBase) containsView(name string) bool {
	_, ok := census.viewNameSet[name]
	return ok
}

func (census *OpenCensusBase) containsMeasure(name string) bool {
	_, ok := census.measureMap[name]
	return ok
}

// Given the censusArgument, initialize the OpenCensus framework
func (census *OpenCensusBase) ViewRegistration(myView *(view.View)) error {
	if census.containsView(myView.Name) == true {
		return errors.Errorf("View has been registered bofore\n")
	}
	if census.containsMeasure(myView.Measure.Name()) == true {
		return errors.Errorf("Measure has been registered bofore\n")
	}
	// The view has never been registered before.
	if err := view.Register(myView); err != nil {
		return err
	} else {
		census.viewNameSet[myView.Name] = 1
		census.measureMap[myView.Measure.Name()] = myView.Measure
		// One Measure Correspond to One []TagKey
		census.tagKeyMap[myView.Measure.Name()] = myView.TagKeys
	}
	// There should be some concurrency control.
	census.status = 1
	return nil
}

func (census *OpenCensusBase) Record(arguments *Protocol.MeasureArgument) error {
	if census.status == 0 {
		return errors.Errorf("Registration Unfinished!")
	}
	measureName := arguments.Name
	if census.containsMeasure(measureName) == false {
		return errors.Errorf("Measurement is not registered")
	} else {
		measure := census.measureMap[measureName]

		// Insert tag values to the context if it exists
		tagKeys := census.tagKeyMap[measureName]
		if len(tagKeys) < len(arguments.TagValues) {
			return errors.Errorf("Number of tag values is more than number of tag keys")
		}
		var mutators []tag.Mutator
		for index, tagValue := range arguments.TagValues {
			if tagValue != "" {
				mutators = append(mutators, tag.Insert(tagKeys[index], tagValue))
			}
		}
		ctx, err := tag.New(census.ctx,
			mutators...,
		)
		if err != nil {
			return err
		}

		switch arguments.MeasureType {
		case "float64":
			if floatMeasure, ok := measure.(*stats.Float64Measure); ok == true {
				value, err := strconv.ParseFloat(arguments.MeasureValue, 64)
				if err != nil {
					return errors.Errorf("Could not Parse the Value: %s because %s",
						arguments.MeasureValue, err.Error())
				} else {
					log.Printf("Record Data %v", value)
					stats.Record(ctx, floatMeasure.M(float64(value)))
				}
			} else {
				return errors.Errorf("Measure Assertion Fails")
			}
		case "int64":
			if intMeasure, ok := measure.(*stats.Int64Measure); ok == true {
				value, err := strconv.ParseFloat(arguments.MeasureValue, 64)
				if err != nil {
					return errors.Errorf("Could not Parse the Value: %s because %s",
						arguments.MeasureValue, err.Error())
				} else {
					log.Printf("Record Data %v", value)
					stats.Record(ctx, intMeasure.M(int64(value)))
				}
			} else {
				return errors.Errorf("Measure Assertion Fails")
			}
		default:
			return errors.Errorf("Unsupported Measure Type")
		}
	}
	return nil
}
