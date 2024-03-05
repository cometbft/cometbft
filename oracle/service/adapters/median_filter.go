package adapters

import (
	"fmt"
	"sort"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
)

// MedianFilter struct for median filter
type MedianFilter struct {
	redisService *redis.Service
}

// Id returns median filter Id
func (filter *MedianFilter) Id() string {
	return "median_filter"
}

func NewMedianFilter(redisService *redis.Service) *MedianFilter {
	return &MedianFilter{
		redisService: redisService,
	}
}

// Validate validate job config
func (filter *MedianFilter) Validate(job types.OracleJob) error {
	valueIds := job.ConfigValue("value_ids").StringArray()
	percentage := job.ConfigValue("tolerance_percentage").Float64()
	switch {
	case len(valueIds) == 0:
		return fmt.Errorf("value_ids cannot be blank")
	case percentage < 0 || percentage > 100:
		return fmt.Errorf("tolerance_percentage must be in range [0, 100]")
	}
	return nil
}

// CalcMedian calculate the median of an array of float64 values
func CalcMedian(values []float64) float64 {
	sort.Float64s(values)
	mid := len(values) / 2

	if len(values)%2 == 1 {
		return values[mid]
	}

	sum := values[mid-1] + values[mid]
	return sum / 2
}

// Perform sets the input value to an empty string if its deviation from the median exceeds the specified threshold
func (filter *MedianFilter) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	input := job.GetInput(result)
	if input.IsEmpty() {
		return result, fmt.Errorf("%s: input cannot be empty", job.InputId)
	}

	valueIds := job.ConfigValue("value_ids").StringArray()
	tolerancePercentage := job.ConfigValue("tolerance_percentage").Float64()
	var values []float64
	for _, valueId := range valueIds {
		value := result.GetData(valueId)
		if value.Present() {
			values = append(values, value.Float64())
		}
	}

	if len(values) == 0 {
		return result, fmt.Errorf("%s: no values found for value_ids: %v", job.InputId, valueIds)
	}

	median := CalcMedian(values)
	output := input
	minValue := median - median/100*tolerancePercentage
	maxValue := median + median/100*tolerancePercentage

	if input.Float64() < minValue || input.Float64() > maxValue {
		output = types.StringToGenericValue("")
	}

	result = job.SetOutput(result, output)

	return result, nil
}
