package adapters

import (
	"fmt"

	"github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/redis"
)

// WeightedAverage struct for weighted average
type WeightedAverage struct {
	redisService *redis.Service
}

func NewWeightedAverage(redisService *redis.Service) *WeightedAverage {
	return &WeightedAverage{
		redisService: redisService,
	}
}

// Id returns weighted average Id
func (adapter *WeightedAverage) Id() string {
	return "weighted_average"
}

// Validate validate job config
func (adapter *WeightedAverage) Validate(job types.OracleJob) error {
	valueIds := job.ConfigValue("value_ids").StringArray()
	weights := job.ConfigValue("weights").Float64Array()
	switch {
	case len(valueIds) != len(weights):
		return fmt.Errorf("len(value_ids) != len(weights)")
	case len(valueIds) == 0:
		return fmt.Errorf("value_ids cannot be 0")
	}
	return nil
}

// CalcWeightedAverage calculate the mean of an array of float64 values
func CalcWeightedAverage(values []float64, weights []float64) float64 {
	var sum float64
	var weightSum float64

	for index, value := range values {
		weight := weights[index]
		if value == 0 || weight == 0 {
			continue
		}
		sum += value * weight
		weightSum += weight
	}

	if weightSum < 0.5 {
		return 0
	}

	return sum / weightSum
}

// Perform calculates the weighted average
func (adapter *WeightedAverage) Perform(job types.OracleJob, result types.AdapterResult, runTimeInput types.AdapterRunTimeInput, _ *types.AdapterStore) (types.AdapterResult, error) {
	valueIds := job.ConfigValue("value_ids").StringArray()
	weights := job.ConfigValue("weights").Float64Array()
	var filteredWeights []float64
	var values []float64
	for index, valueId := range valueIds {
		value := result.GetData(valueId)
		if value.Present() {
			values = append(values, value.Float64())
			filteredWeights = append(filteredWeights, weights[index])
		}
	}

	if len(values) == 0 {
		return result, fmt.Errorf("no values found for weighted_average")
	}

	average := CalcWeightedAverage(values, filteredWeights)

	if average == 0 {
		return result, fmt.Errorf("weighted_average is 0")
	}

	result = job.SetOutput(result, types.Float64ToGenericValue(average))

	return result, nil
}
