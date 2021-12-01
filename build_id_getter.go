package main

import (
	"path/filepath"

	"github.com/bitrise-steplib/bitrise-step-artifact-pull/model"
)

const DELIMITER = "."

type BuildIDGetter interface {
	GetBuildIDs() ([]string, error)
}

type DefaultBuildIDGetter struct {
	FinishedStages model.FinishedStages
	TargetNames    []string
}

func NewDefaultBuildIDGetter(finishedStages model.FinishedStages, targetNames []string) BuildIDGetter {
	return DefaultBuildIDGetter{
		FinishedStages: finishedStages,
		TargetNames:    targetNames,
	}
}

func (bg DefaultBuildIDGetter) GetBuildIDs() ([]string, error) {
	var buildIDs []string

	stageWorkflowMap := bg.createWorkflowMap()

	if len(bg.TargetNames) == 0 {
		for _, v := range stageWorkflowMap {
			buildIDs = append(buildIDs, v)
		}

		return buildIDs, nil
	}

	for _, target := range bg.TargetNames {
		for k, v := range stageWorkflowMap {
			matched, err := filepath.Match(target, k)
			if err != nil {
				return nil, err
			}

			if matched {
				buildIDs = append(buildIDs, v)
			}
		}
	}

	return buildIDs, nil
}

func (bg DefaultBuildIDGetter) createWorkflowMap() map[string]string {
	stageWorkflowMap := map[string]string{}
	for _, stage := range bg.FinishedStages.Stages {
		for _, wf := range stage.Workflows {
			stageWorkflowMap[stage.Name+DELIMITER+wf.Name] = wf.ExternalId
		}
	}

	return stageWorkflowMap
}
