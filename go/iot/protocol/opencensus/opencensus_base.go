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
	"fmt"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/protocol"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

type OpenCensusBase struct {
	ctx                context.Context
	registeredMeasures map[string]stats.Measure // Store all the measure based on their Name. Used for the future record
	// TODO: What if different views share the same tag key
	registeredTagKeys map[string]tag.Key
}

func (census *OpenCensusBase) Initialize(projectId string, reportPeriod int) {
	census.ctx = context.Background()
	census.registeredMeasures = make(map[string]stats.Measure)
	census.registeredTagKeys = make(map[string]tag.Key)
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
	_, ok := census.registeredMeasures[name]
	return ok
}

func (census *OpenCensusBase) isMeasureConflict(measure *stats.Measure) bool {
	var myMeasure = census.registeredMeasures[(*measure).Name()]
	if myMeasure.Description() != (*measure).Description() || myMeasure.Unit() != (*measure).Unit() {
		return true
	}
	return false
}

// Given the censusArgument, initialize the OpenCensus framework
func (census *OpenCensusBase) ViewRegistration(myView *(view.View)) error {
	// The view has never been registered before.
	if err := view.Register(myView); err != nil {
		return err
	} else {
		if census.containsMeasure(myView.Measure.Name()) {
			if flag := census.isMeasureConflict(&myView.Measure); flag == true {
				return errors.Errorf("Different measures share the same name!")
			}
		} else {
			census.registeredMeasures[myView.Measure.Name()] = myView.Measure
		}
		// Store the tag name
		var tagKeys = myView.TagKeys
		for _, key := range tagKeys {
			census.registeredTagKeys[key.Name()] = key
		}
	}
	return nil
}

func (census *OpenCensusBase) insertTag(tagPairs map[string]string) (context.Context, bool, error) {
	// Insert tag values to the context if it exists
	// Normally the program returns the context and nil error
	// But when any tag key doesn't exist, we still return the context but don't insert that tag key
	var mutators []tag.Mutator
	var tagExist = true
	for key, value := range tagPairs {
		tagKey, ok := census.registeredTagKeys[key]
		if ok == true {
			// The tag key exists
			mutators = append(mutators, tag.Insert(tagKey, value))
		} else {
			tagExist = false
		}
	}
	ctx, err := tag.New(census.ctx,
		mutators...,
	)
	return ctx, tagExist, err
}
func (census *OpenCensusBase) Record(arguments *protocol.MeasureArgument) *protocol.Response {
	measureName := arguments.Name
	if census.containsMeasure(measureName) == false {
		return &protocol.Response{protocol.UNREGISTERMEASURE, "Measure is not registered"}
	} else {
		measure := census.registeredMeasures[measureName]

		ctx, tagExist, err := census.insertTag(arguments.Tags)

		if err != nil {
			return &protocol.Response{protocol.FAIL, err.Error()}
		}

		if value, err := strconv.ParseFloat(arguments.Value, 64); err != nil {
			info := fmt.Sprintf("Could not parse the value: %s because %s", arguments.Value, err.Error())
			return &protocol.Response{protocol.FAIL, info}
		} else {
			//log.Printf("Record Data %v", value)
			switch vv := measure.(type) {
			case *stats.Float64Measure:
				stats.Record(ctx, vv.M(float64(value)))
				break
			case *stats.Int64Measure:
				stats.Record(ctx, vv.M(int64(value)))
				break
			default:
				return &protocol.Response{protocol.FAIL, "Unsupported measure type"}
			}
		}

		if tagExist {
			return &protocol.Response{protocol.OK, ""}
		} else {
			return &protocol.Response{protocol.UNREGISTERTAG, "Tags key doesn't exist."}
		}
	}
}
