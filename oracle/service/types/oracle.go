package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cometbft/cometbft/redis"
)

// Oracle struct for oracles
type Oracle struct {
	Id         string     `json:"id"`
	Resolution uint64     `json:"resolution"`
	Spec       OracleSpec `json:"spec"`
}

// OracleSpec struct for oracle specs
type OracleSpec struct {
	OutputId  string           `json:"output_id"`
	Jobs      []OracleJob      `json:"jobs"`
	Templates []OracleTemplate `json:"templates"`
	// flag to determine if oracle jobs should terminate upon error (should terminate for specs that do not use weighted_average in their calcs.)
	ShouldEarlyTerminate bool `json:"should_early_terminate"`
}

// OracleJob struct for oracle jobs
type OracleJob struct {
	OutputId    string                `json:"output_id"`
	InputId     string                `json:"input_id"`
	Adapter     string                `json:"adapter"`
	Config      map[string]string     `json:"config"`
	SubAdapters []OracleJobSubAdapter `json:"adapters"`
}

// OracleTemplate struct for oracle template
type OracleTemplate struct {
	TemplateId  string                `json:"template_id"`
	SubAdapters []OracleJobSubAdapter `json:"adapters"`
}

// OracleJobSubAdapter struct for sub adapters
type OracleJobSubAdapter struct {
	Adapter  string            `json:"adapter"`
	Config   map[string]string `json:"config"`
	Template string            `json:"template"`
}

type JSONTime struct {
	Time time.Time
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.Time.Format(time.RFC3339Nano))), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	var timestamp string
	err := json.Unmarshal(data, &timestamp)
	if err != nil {
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

type OracleCache struct {
	Price     string   `json:"price"`
	Timestamp JSONTime `json:"timestamp"`
}

// ConfigValue returns the config for a specified key
func (job OracleJob) ConfigValue(key string) GenericValue {
	return StringToGenericValue(job.Config[key])
}

// GetInput gets job input for the specified adapter result
func (job OracleJob) GetInput(result AdapterResult) GenericValue {
	return result.GetData(job.InputId)
}

// GetInputs gets list of job input for the specified adapter result
func (job OracleJob) GetInputs(result AdapterResult) ([]GenericValue, error) {
	res := []GenericValue{}
	input := result.GetData(job.InputId)
	if input.IsEmpty() {
		valueIdsStr := job.ConfigValue("input_ids")
		if !valueIdsStr.Present() || valueIdsStr.IsEmpty() {
			return res, fmt.Errorf("GetInputs: input_ids not found in job config for job with outputID: %s", job.OutputId)
		}
		valueIds := valueIdsStr.StringArray()
		for _, valueId := range valueIds {
			res = append(res, result.GetData(valueId))
		}
	} else {
		values := strings.Split(input.String(), " ")
		for _, val := range values {
			res = append(res, redis.StringToGenericValue(val))
		}
	}
	return res, nil
}

// SetOutput sets the job output on the adapter result
func (job OracleJob) SetOutput(result AdapterResult, output GenericValue) AdapterResult {
	result.SetData(job.OutputId, output)
	return result
}

// SetOutputList sets the job output of an array of strings on the adapter result
func (job OracleJob) SetOutputList(result AdapterResult, output []string) AdapterResult {
	job.SetOutput(result, StringArrayToGenericValue(output))
	return result
}

// SetOutputs sets the job outputs on the adapter results
func (job OracleJob) SetOutputs(result AdapterResult, outputIds []string, outputs []GenericValue) AdapterResult {
	for idx := range outputIds {
		result.SetData(outputIds[idx], outputs[idx])
	}
	return result
}

func StringArrayToGenericValue(stringArray []string) GenericValue {
	strings := strings.Join(stringArray, " ")
	return StringToGenericValue(strings)
}
