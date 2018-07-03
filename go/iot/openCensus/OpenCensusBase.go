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
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/Protocol"
	"github.com/mgutz/logxi/v1"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	FLOAT = 0
	INT   = 1
)

// TODO: Name of this struct ??
type OpenCensusBase struct {
	//TODO: Should I implement state machine pattern or simply use flag?
	status       int
	ctx          context.Context
	projectIdSet map[string]int
	// Use the name of View as the index
	// Since we might add tagKey to the view in the future. It's better to use pointers.
	viewSet map[string]*(view.View)
	// TODO: Look like that the report Period for different views is the same
	reportPeriod int
	// TODO: Do we need to handle the concurrent problems?
	// Store all the measure based on their Name. Used for the future record
	measureMap map[string]stats.Measure

	tagKeyMap map[string]tag.Key
}

// TODO: There might be better ways to initialize the map
func (census *OpenCensusBase) Initialize() {
	census.status = 0
	census.ctx = context.Background()
	census.projectIdSet = make(map[string]int)
	census.viewSet = make(map[string]*(view.View))
	census.measureMap = make(map[string]stats.Measure)
	census.tagKeyMap = make(map[string]tag.Key)
}

// Return whether the view has already been registered
func (census *OpenCensusBase) containsView(name string) bool {
	_, ok := census.viewSet[name]
	return ok
}

// Return whether the project Id has already been registered
func (census *OpenCensusBase) containsProjId(name string) bool {
	_, ok := census.projectIdSet[name]
	return ok
}

func (census *OpenCensusBase) containsMeasure(name string) bool {
	_, ok := census.measureMap[name]
	return ok
}

// Given the censusArgument, initialize the OpenCensus framework
func (census *OpenCensusBase) InitOpenCensus(arguments *Protocol.Argument) error {
	// TODO: Currently we assume for each arduino there would be only one view.
	if census.status == 1 {
		return errors.Errorf("View Already Been Registered!")
	}
	// Register Exporter if necessary
	projectId := arguments.ProjectId
	if census.containsProjId(projectId) == false {
		// The exporter has never been registered before.
		// Create a new stackdriver exporter and register it.
		if exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: projectId,
		}); err != nil {
			return err
		} else {
			// TODO: It's better to have err return here. But openCensus doesn't support yet.
			view.RegisterExporter(exporter)
			census.projectIdSet[projectId] = 1
		}
	} else {
		// TODO: Log or directly return error?
		return errors.Errorf("Exporter already been registered!")
	}

	// Register view if necessary
	viewInput := arguments.View
	if err := Protocol.ViewParse(&viewInput, arguments); err != nil {
		return err
	}

	// TODO:is View implemented by the raspberry pi or Arduino ????
	if census.containsView(viewInput.Name) == false {
		// The view has never been registered before.
		// Create a new view and register it.

		if err := view.Register(&viewInput); err != nil {
			return err
		} else {
			// TODO： Never check whether the tagKey exists in the map before?
			for _, tagKey := range viewInput.TagKeys {
				census.tagKeyMap[tagKey.Name()] = tagKey
			}
			census.viewSet[viewInput.Name] = &viewInput
			// TODO: Temporarily combine the measurement and view
			census.measureMap[viewInput.Measure.Name()] = viewInput.Measure
		}
	} else {
		if err := census.checkViewConflict(arguments); err != nil {
			return err
		} else {
			viewInput := arguments.View
			viewRegistered := census.viewSet[viewInput.Name]
			appendTags(viewRegistered, viewInput)
		}
	}
	// Set reporting period to report data
	view.SetReportingPeriod(time.Second * time.Duration(arguments.ReportPeriod))
	census.reportPeriod = arguments.ReportPeriod
	census.status = 1
	return nil
}

func appendTags(viewRegistered *view.View, viewInput view.View) {
	for _, newKey := range viewInput.TagKeys {
		// TODO: Not check existence yet
		viewRegistered.TagKeys = append(viewRegistered.TagKeys, newKey)
	}
}
func (census *OpenCensusBase) checkViewConflict(argument *(Protocol.Argument)) error {
	viewInput := argument.View
	viewRegistered := census.viewSet[viewInput.Name]
	if viewRegistered.Description != viewInput.Description {
		return errors.Errorf("Conflict Description for View")
	} else if viewRegistered.Measure.Name() != viewInput.Measure.Name() ||
		viewRegistered.Measure.Description() != viewInput.Measure.Description() ||
		viewRegistered.Measure.Unit() != viewRegistered.Measure.Unit() {
		return errors.Errorf("Conflict Measure for View")
	}
	return nil

}
func (census *OpenCensusBase) Record(arguments *Protocol.Argument) error {
	measureName := arguments.Measure.Name
	if census.containsMeasure(measureName) == false {
		return errors.Errorf("Measurement already been registered")
	} else {
		measure := census.measureMap[measureName]
		// TODO: Assume that no conflict between the initial measure type and later one
		switch arguments.Measure.MeasureType {
		case "float64":
			// TODO: Assertion Bug???
			floatMeasure, ok := measure.(*stats.Float64Measure)
			if ok == true {
				// TODO: Do we need to check assertion?
				value, err := strconv.ParseFloat(arguments.Measure.MeasureValue, 64)
				if err != nil {
					return errors.Errorf("Could not Parse the Value: %s because %s",
						arguments.Measure.MeasureValue, err.Error())
				} else {
					log.Info("Record Data %v", value)
					stats.Record(census.ctx, floatMeasure.M(float64(value)))
				}
			} else {
				return errors.Errorf("Measure Assertion Fails")
			}
		case "int64":
			intMeasure, ok := measure.(*stats.Int64Measure)
			if ok == true {
				// TODO: Do we need to check assertion?
				value, err := strconv.ParseFloat(arguments.Measure.MeasureValue, 64)
				if err != nil {
					return errors.Errorf("Could not Parse the Value: %s because %s",
						arguments.Measure.MeasureValue, err.Error())
				} else {
					log.Info("Record Data %v", value)
					stats.Record(census.ctx, intMeasure.M(int64(value)))
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
