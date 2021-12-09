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

func NewBuildIDGetter(finishedStages model.FinishedStages, targetNames []string) BuildIDGetter {
	return BuildIDGetter{
		FinishedStages: finishedStages,
		TargetNames:    targetNames,
	}
}

func (bg BuildIDGetter) GetBuildIDs() ([]string, error) {
	buildIDsSet := make(map[string]bool)

	stageWorkflowMap := bg.createWorkflowMap()

	if len(bg.TargetNames) == 0 {
		for _, v := range stageWorkflowMap {
			buildIDsSet[v] = true
		}

		return convertKeySetToArray(buildIDsSet), nil
	}

	for _, target := range bg.TargetNames {
		for k, v := range stageWorkflowMap {
			matched, err := filepath.Match(target, k)
			if err != nil {
				return nil, err
			}

			if matched {
				buildIDsSet[v] = true
			}
		}
	}

	return convertKeySetToArray(buildIDsSet), nil
}

func convertKeySetToArray(set map[string]bool) []string {
	var ids []string

	for k := range set {
		ids = append(ids, k)
	}

	return ids
}

func (bg BuildIDGetter) createWorkflowMap() map[string]string {
	stageWorkflowMap := map[string]string{}
	for _, stage := range bg.FinishedStages {
		for _, wf := range stage.Workflows {
			stageWorkflowMap[stage.Name+DELIMITER+wf.Name] = wf.ExternalId
		}
	}

	return stageWorkflowMap
}
