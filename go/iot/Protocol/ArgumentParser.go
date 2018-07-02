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

package Protocol

import (
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	v "go.opencensus.io/stats/view"
)

func ViewParse(view *view.View, argument *Argument) error {
	// Name and Description of the view should be already updated by the JSON.
	// TODO: If not using JSON, we should update the name and description by ourseleves
	measureString := argument.Measure
	switch measureString.MeasureType {
	case "int64":
		view.Measure = stats.Int64(measureString.Name, measureString.Description, measureString.Unit)
	case "float64":
		view.Measure = stats.Float64(measureString.Name, measureString.Description, measureString.Unit)
	default:
		return errors.Errorf("Unsupported Measure Type")
	}

	aggregationString := argument.Aggregation
	switch aggregationString.AggregationType {
	case "LastValue":
		view.Aggregation = v.LastValue()
	case "Sum":
		view.Aggregation = v.Sum()
	case "Count":
		view.Aggregation = v.Count()
	case "Distribution":
		// TODO: Don't know how to transfer from Slice to ... argument
		view.Aggregation = v.Distribution(aggregationString.AggregationValue[0])
		//view.Aggregation = v.Distribution(aggregationString.AggregationValue[0], aggregationString.AggregationValue[1:]...)
	default:
		return errors.Errorf("Unsupported Aggregation Type")
	}
	return nil
}
