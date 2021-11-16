package main

import (
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
	"testing"

	mockenv "github.com/bitrise-io/go-utils/env/mocks"
)

func Test_GivenInputs_WhenCreatingConfig_ThenMappingIsCorrect(t *testing.T) {
	// Given
	envRepository := new(mockenv.Repository)
	envRepository.On("Get", "verbose").Return("true")
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
	assert.Equal(t, true, config.Verbose)
}
