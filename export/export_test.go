package export

import (
	"errors"
	"testing"

	mockenv "github.com/bitrise-io/go-utils/env/mocks"
	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
)

func TestExport_NoError(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "ENV_KEY", "a/b.txt").Return(nil)

	exporter := OutputExporter{
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
	}

	err := exporter.Export(map[string]string{"ENV_KEY": "a/b.txt"})
	assert.NoError(t, err)

	envRepository.AssertExpectations(t)
}

func TestExport_Error(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "ENV_KEY", "a/b.txt").Return(errors.New("some kind of error"))

	exporter := OutputExporter{
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
	}

	err := exporter.Export(map[string]string{"ENV_KEY": "a/b.txt"})
	assert.EqualError(t, err, "failed to export pulled artifact location (ENV_KEY=a/b.txt), error: some kind of error")

	envRepository.AssertExpectations(t)
}
