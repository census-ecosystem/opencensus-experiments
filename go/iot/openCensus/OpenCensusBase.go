package openCensus

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/census-ecosystem/opencensus-experiments/go/iot/Protocol"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"log"
	"strconv"
	"time"
)

const (
	FLOAT = 0
	INT   = 1
)

// TODO: Name of this struct ??
type OpenCensusBase struct {
	//TODO: Delete it in the future
	status       int
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

// TODO: There might be better ways to initialize the map
func (census *OpenCensusBase) Initialize() {
	census.status = 0
	census.ctx = context.Background()
	census.projectIdSet = make(map[string]int)
	census.viewSet = make(map[string]view.View)
	census.measureMap = make(map[string]stats.Measure)
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
	if census.status == 1 {
		log.Println("Already Registered!\n")
		return nil
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
		log.Println("Exporter already exists\n")
	}

	// Register view if necessary
	viewInput := arguments.View
	Protocol.ViewParse(&viewInput, arguments)

	if census.containsView(viewInput.Name) == false {
		// The view has never been registered before.
		// Create a new view and register it.

		if err := view.Register(&viewInput); err != nil {
			return err
		} else {
			census.viewSet[viewInput.Name] = viewInput
			// TODO: Temporarily combine the measurement and view
			census.measureMap[viewInput.Measure.Name()] = viewInput.Measure
		}
	} else {
		// The view has already been registered before. We don't need to register again
		log.Println("View already exists\n")
	}
	// Set reporting period to report data
	view.SetReportingPeriod(time.Second * time.Duration(arguments.ReportPeriod))
	census.reportPeriod = arguments.ReportPeriod
	census.status = 1
	log.Println("fUCK")
	return nil
}

func (census *OpenCensusBase) Record(arguments *Protocol.Argument) error {
	measureName := arguments.Measure.Name
	if census.containsMeasure(measureName) == false {
		return errors.Errorf("The Measurement has never been registered\n")
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
					log.Printf("Could Parse the Value: %s\n", arguments.Measure.MeasureValue)
				} else {
					stats.Record(census.ctx, floatMeasure.M(float64(value)))
				}
			} else {
				return errors.Errorf("The Measure Assertion Fails\n")
			}
		case "int64":
			intMeasure, ok := measure.(*stats.Int64Measure)
			if ok == true {
				// TODO: Do we need to check assertion?
				value, err := strconv.ParseFloat(arguments.Measure.MeasureValue, 64)
				if err != nil {
					log.Printf("Could Parse the Value: %s\n", arguments.Measure.MeasureValue)
				} else {
					stats.Record(census.ctx, intMeasure.M(int64(value)))
				}
			} else {
				return errors.Errorf("The Measure Assertion Fails\n")
			}
		default:
			return errors.Errorf("No Such Kind Errors\n")
		}
	}
	return nil
}
