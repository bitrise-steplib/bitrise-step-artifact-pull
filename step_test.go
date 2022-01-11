package main

import (
	"testing"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	mockenv "github.com/bitrise-io/go-utils/env/mocks"
	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
)

func Test_GivenInputs_WhenCreatingConfig_ThenMappingIsCorrect(t *testing.T) {
	// Given
	envRepository := new(mockenv.Repository)
	envRepository.On("Get", "BITRISE_APP_SLUG").Return("app-slug")
	envRepository.On("Get", "verbose").Return("true")
	envRepository.On("Get", "artifact_sources").Return("stage1.workflow1,stage2.*")
	envRepository.On("Get", "finished_stage").Return("")
	envRepository.On("Get", "bitrise_api_base_url").Return("")
	envRepository.On("Get", "bitrise_api_access_token").Return("")
	inputParser := stepconf.NewInputParser(envRepository)
	cmdFactory := command.NewFactory(envRepository)
	step := ArtifactPull{
		inputParser:   inputParser,
		envRepository: envRepository,
		cmdFactory:    cmdFactory,
		logger:        log.NewLogger(),
	}

	// When
	config, err := step.ProcessConfig()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, true, config.VerboseLogging)
	assert.Equal(t, []string{"stage1.workflow1", "stage2.*"}, config.ArtifactSources)
}

func Test_Export(t *testing.T) {
	envRepository := new(mockenv.Repository)

	testCases := []struct {
		desc                string
		inputResult         Result
		expectedExportValue string
	}{
		{
			desc: "when there are more than one result, it exports a coma separated list",
			inputResult: Result{
				ArtifactLocations: []string{"aa.txt", "bb.txt"},
			},
			expectedExportValue: "aa.txt|bb.txt",
		},
		{
			desc: "when there is a result element",
			inputResult: Result{
				ArtifactLocations: []string{"aa.txt"},
			},
			expectedExportValue: "aa.txt",
		},
		{
			desc: "when there is no result element",
			inputResult: Result{
				ArtifactLocations: []string{},
			},
			expectedExportValue: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			step := ArtifactPull{
				envRepository: envRepository,
				logger:        log.NewLogger(),
			}

			envRepository.On("Set", "BITRISE_ARTIFACT_PATHS", tC.expectedExportValue).Return(nil)

			err := step.Export(tC.inputResult, make(map[string]string))

			envRepository.AssertExpectations(t)
			assert.NoError(t, err)
		})
	}
}
