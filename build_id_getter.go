package main

import (
	"path/filepath"

	"github.com/bitrise-steplib/bitrise-step-artifact-pull/model"
)

const DELIMITER = "."

type BuildIDGetter struct {
	FinishedStages model.FinishedStages
	TargetNames    []string
}

type keyValuePair struct {
	key   string
	value string
}

func NewBuildIDGetter(finishedStages model.FinishedStages, targetNames []string) BuildIDGetter {
	return BuildIDGetter{
		FinishedStages: finishedStages,
		TargetNames:    targetNames,
	}
}

func (bg BuildIDGetter) GetBuildIDs() ([]string, error) {
	var buildIDs []string

	kvpSlice := bg.createKeyValuePairSlice()

	if len(bg.TargetNames) == 0 {
		for _, kvPair := range kvpSlice {
			buildIDs = append(buildIDs, kvPair.value)
		}

		return buildIDs, nil
	}

	for _, target := range bg.TargetNames {
		for _, kvPair := range kvpSlice {
			matched, err := filepath.Match(target, kvPair.key)
			if err != nil {
				return nil, err
			}

			if matched {
				buildIDs = append(buildIDs, kvPair.value)
			}
		}
	}

	return buildIDs, nil
}

func (bg BuildIDGetter) createKeyValuePairSlice() []keyValuePair {
	var stageWorkflowMap []keyValuePair
	for _, stage := range bg.FinishedStages {
		for _, wf := range stage.Workflows {
			stageWorkflowMap = append(stageWorkflowMap, keyValuePair{
				key:   stage.Name + DELIMITER + wf.Name,
				value: wf.ExternalId,
			})
		}
	}

	return stageWorkflowMap
}
