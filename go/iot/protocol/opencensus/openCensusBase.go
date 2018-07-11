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

package opencensus

import (
	"context"
	"log"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
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
	ctx        context.Context
	measureMap map[string]stats.Measure // Store all the measure based on their Name. Used for the future record
	// TODO: What if different views share the same tag key
	tagKeyMap map[string]tag.Key
}

func (census *OpenCensusBase) Initialize(projectId string, reportPeriod int) {
	census.ctx = context.Background()
	census.measureMap = make(map[string]stats.Measure)
	census.tagKeyMap = make(map[string]tag.Key)
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectId, // Google Cloud Console project ID.
	})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(exporter)
	view.SetReportingPeriod(time.Second * time.Duration(reportPeriod))
}

func (census *OpenCensusBase) containsMeasure(name string) bool {
	_, ok := census.measureMap[name]
	return ok
}

// Given the censusArgument, initialize the OpenCensus framework
func (census *OpenCensusBase) ViewRegistration(myView *(view.View)) error {
	if census.containsMeasure(myView.Measure.Name()) == true {
		return errors.Errorf("Measure has been registered bofore\n")
	}
	// The view has never been registered before.
	if err := view.Register(myView); err != nil {
		return err
	} else {
		census.measureMap[myView.Measure.Name()] = myView.Measure
		// Store the tag name
		var tagKeys = myView.TagKeys
		for _, key := range tagKeys {
			census.tagKeyMap[key.Name()] = key
		}
	}
	return nil
}

func (census *OpenCensusBase) insertTag(tagPairs map[string]interface{}) (context.Context, bool, error) {
	// Insert tag values to the context if it exists
	// Normally the program returns the context and nil error
	// But when any tag key doesn't exist, we still return the context but don't insert that tag key
	var mutators []tag.Mutator
	var tagExist = true
	for key, value := range tagPairs {
		tagKey, ok := census.tagKeyMap[key]
		if ok == true {
			// The tag key exists
			// TODO: doesn't check the type of the value
			mutators = append(mutators, tag.Insert(tagKey, value.(string)))
		} else {
			tagExist = false
		}
	}
	ctx, err := tag.New(census.ctx,
		mutators...,
	)
	return ctx, tagExist, err
}
func (census *OpenCensusBase) Record(arguments *protocol.MeasureArgument) (int, error) {
	measureName := arguments.Name
	if census.containsMeasure(measureName) == false {
		return protocol.UNREGISTERMEASURE, errors.Errorf("Measurement is not registered")
	} else {
		measure := census.measureMap[measureName]

		ctx, tagExist, err := census.insertTag(arguments.Tag)

		if err != nil {
			return protocol.FAIL, nil
		}

		if value, err := strconv.ParseFloat(arguments.Measurement, 64); err != nil {
			return protocol.FAIL, errors.Errorf("Could not Parse the Value: %s because %s",
				arguments.Measurement, err.Error())
		} else {
			//log.Printf("Record Data %v", value)
			switch vv := measure.(type) {
			case *stats.Float64Measure:
				stats.Record(ctx, vv.M(float64(value)))
			case *stats.Int64Measure:
				stats.Record(ctx, vv.M(int64(value)))
			default:
				return protocol.FAIL, errors.Errorf("Unsupported Measure Type")
			}
		}

		if tagExist {
			return protocol.OK, nil
		} else {
			return protocol.UNREGISTERTAG, errors.Errorf("Tag key doesn't exist.")
		}
	}
}
