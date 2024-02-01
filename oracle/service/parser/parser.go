package parser

import (
	"fmt"

	"github.com/cometbft/cometbft/oracle/service/types"
)

// ConstructTemplateMap create a map of templates
func ConstructTemplateMap(templates []types.OracleTemplate) (map[string]types.OracleTemplate, error) {
	templateMap := make(map[string]types.OracleTemplate, len(templates))

	for _, template := range templates {
		if _, ok := templateMap[template.TemplateId]; ok {
			return nil, fmt.Errorf("duplicate template id: %s", template.TemplateId)
		}
		templateMap[template.TemplateId] = template
	}
	return templateMap, nil
}

// UnrollSubAdapters convert sub adapter templates to full sub adapters
func UnrollSubAdapters(subAdapters []types.OracleJobSubAdapter, templateMap map[string]types.OracleTemplate) ([]types.OracleJobSubAdapter, error) {
	var unrolledAdapters []types.OracleJobSubAdapter
	for _, subAdapter := range subAdapters {
		switch {
		case len(subAdapter.Adapter) > 0:
			unrolledAdapters = append(unrolledAdapters, subAdapter)

		case len(subAdapter.Template) > 0:
			template, ok := templateMap[subAdapter.Template]
			if !ok {
				return nil, fmt.Errorf("template not found in map: %s", subAdapter.Template)
			}
			unrolledAdapters = append(unrolledAdapters, template.SubAdapters...)

		default:
			return nil, fmt.Errorf("invalid subadapter")
		}
	}
	return unrolledAdapters, nil
}

// UnrollOracleJobs convert sub adapter declarations to full jobs
func UnrollOracleJobs(jobs []types.OracleJob) (unrolledJobs []types.OracleJob, err error) {
	for _, job := range jobs {
		switch {
		case len(job.Adapter) > 0:
			unrolledJobs = append(unrolledJobs, job)

		case len(job.SubAdapters) > 0:
			for _, subAdapter := range job.SubAdapters {
				unrolledJob := types.OracleJob{}
				unrolledJob.OutputId = job.OutputId
				unrolledJob.InputId = job.InputId
				if len(unrolledJob.InputId) == 0 {
					unrolledJob.InputId = job.OutputId
				}
				unrolledJob.Adapter = subAdapter.Adapter
				unrolledJob.Config = subAdapter.Config
				unrolledJobs = append(unrolledJobs, unrolledJob)
			}
		default:
			return nil, fmt.Errorf("job has no adapters or subadapters")
		}
	}
	return
}

func ParseSpec(spec types.OracleSpec) (unrolledSpec types.OracleSpec, err error) {
	templateMap, err := ConstructTemplateMap(spec.Templates)
	if err != nil {
		return
	}
	for i, job := range spec.Jobs {
		spec.Jobs[i].SubAdapters, err = UnrollSubAdapters(job.SubAdapters, templateMap)
		if err != nil {
			return
		}
	}
	spec.Jobs, err = UnrollOracleJobs(spec.Jobs)
	if err != nil {
		return
	}
	return spec, nil
}

// ValidateOracleJobs validates oracle jobs
func ValidateOracleJobs(app types.App, jobs []types.OracleJob) error {
	for _, job := range jobs {
		adapter, ok := app.AdapterMap[job.Adapter]
		if !ok {
			return fmt.Errorf("adapter not found: '%s'", job.Adapter)
		}
		err := adapter.Validate(job)
		if err != nil {
			return fmt.Errorf("%s: %s", adapter.Id(), err.Error())
		}
	}
	return nil
}
