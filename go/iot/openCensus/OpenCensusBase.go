package openCensus

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"log"
	"time"
)

const (
	FLOAT = 0
	INT   = 1
)

// TODO: Name of this struct ??
type OpenCensusBase struct {
	ctx          context.Context
	projectIdSet map[string]int
	// TODO: Use the name of View as the index
	viewSet map[string]view.View
	// TODO: Look like that for different views the report Period is the same
	reportPeriod int
	// TODO: Do we need to handle the concurrent problems?
	// Store all the measure based on their Name. Used for the future record
	measureMap map[string]stats.Measure
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
func (census *OpenCensusBase) InitOpenCensus(arguments *RegisterArgument) error {
	// Register Exporter if necessary
	projectId := arguments.projectId
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
		log.Println("Exporter already exists\n")
	}

	// Register view if necessary
	viewInput := arguments.view
	if census.containsView(viewInput.Name) == false {
		// The view has never been registered before.
		// Create a new view and register it.
		if err := view.Register(&viewInput); err != nil {
			return err
		} else {
			census.viewSet[viewInput.Name] = viewInput
		}
	} else {
		// The view has already been registered before. We don't need to register again
		log.Println("View already exists\n")
	}
	// Set reporting period to report data
	view.SetReportingPeriod(time.Second * time.Duration(arguments.reportPeriod))
	return nil
}

func (census *OpenCensusBase) Record(arguments *RecordArgument) error {
	measureName := arguments.measureName
	if census.containsMeasure(measureName) == false {
		return errors.Errorf("The Measurement has never been registered\n")
	} else {
		measure := census.measureMap[measureName]
		// TODO: Assume that no conflict between the initial measure type and later one
		switch arguments.measureType {
		case 0:
			// TODO: Assertion Bug???
			floatMeasure, ok := measure.(*stats.Float64Measure)
			if ok == true {
				// TODO: Do we need to check assertion?
				stats.Record(census.ctx, floatMeasure.M(arguments.value.(float64)))
			} else {
				return errors.Errorf("The Measure Assertion Fails\n")
			}
		case 1:
			floatMeasure, ok := measure.(*stats.Int64Measure)
			if ok == true {
				// TODO: Do we need to check assertion?
				stats.Record(census.ctx, floatMeasure.M(arguments.value.(int64)))
			} else {
				return errors.Errorf("The Measure Assertion Fails\n")
			}
		default:
			return errors.Errorf("No Such Kind Errors\n")
		}
	}
	return nil
}

type RegisterArgument struct {
	projectId    string
	view         view.View
	reportPeriod int
}

type RecordArgument struct {
	measureName string
	value       interface{}
	measureType int
}

