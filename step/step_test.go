package step

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
	envRepository.On("Get", "artifact_sources").Return("stage1.workflow1,stage2.*")
	envRepository.On("Get", "verbose").Return("true")
	envRepository.On("Get", "app_slug").Return("app-slug")
	envRepository.On("Get", "finished_stage").Return("[]")
	envRepository.On("Get", "bitrise_api_base_url").Return("url")
	envRepository.On("Get", "bitrise_api_access_token").Return("token")

	inputParser := stepconf.NewInputParser(envRepository)
	cmdFactory := command.NewFactory(envRepository)
	step := IntermediateFileDownloader{
		inputParser:   inputParser,
		envRepository: envRepository,
		cmdFactory:    cmdFactory,
		logger:        log.NewLogger(),
	}

	// When
	config, err := step.ProcessConfig()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, []string{"stage1.workflow1", "stage2.*"}, config.ArtifactSources)
}

func Test_Export(t *testing.T) {
	envRepository := new(mockenv.Repository)

	testCases := []struct {
		desc        string
		inputResult Result
	}{
		{
			desc: "when there are more than one result, it exports a coma separated list",
			inputResult: Result{
				IntermediateFiles: map[string]string{"ENV_KEY_A": "aa.txt", "ENV_KEY_B": "bb.txt"},
			},
		},
		{
			desc: "when there is a result element",
			inputResult: Result{
				IntermediateFiles: map[string]string{"ENV_KEY_A": "aa.txt"},
			},
		},
		{
			desc: "when there is no result element",
			inputResult: Result{
				IntermediateFiles: map[string]string{},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			step := IntermediateFileDownloader{
				envRepository: envRepository,
				logger:        log.NewLogger(),
			}

			for envKey, path := range tC.inputResult.IntermediateFiles {
				envRepository.On("Set", envKey, path).Return(nil)
			}

			err := step.Export(tC.inputResult)

			envRepository.AssertExpectations(t)
			assert.NoError(t, err)
		})
	}
}
