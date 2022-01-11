package export

import (
	"errors"
	"testing"

	mockenv "github.com/bitrise-io/go-utils/env/mocks"
	"github.com/bitrise-io/go-utils/log"
	"github.com/stretchr/testify/assert"
)

func TestProcessRawExportMap(t *testing.T) {
	testCases := []struct {
		desc           string
		input          string
		expectedOutput map[string]string
	}{
		{
			desc: "when input is given, parses the input string",
			input: `
DOWNLOADED_APKS - *.apk
DOWNLOADED_TEST_RESULTS - *.result
`,
			expectedOutput: map[string]string{
				"DOWNLOADED_APKS":         "*.apk",
				"DOWNLOADED_TEST_RESULTS": "*.result",
			},
		},
		{
			desc: "when multiple expressions presents in the expression field",
			input: `
ARTIFACTS - *.apk,*ipa
TEXTS - *.txt,*docx
`,
			expectedOutput: map[string]string{
				"ARTIFACTS": "*.apk,*ipa",
				"TEXTS":     "*.txt,*docx",
			},
		},
		{
			desc: "when the input has a messed up line",
			input: `
DOWNLOADED_APKS - *.apk
- *.ipa
DOWNLOADED_TEST_RESULTS - *.result
EMPTY - 
`,
			expectedOutput: map[string]string{
				"DOWNLOADED_APKS":         "*.apk",
				"DOWNLOADED_TEST_RESULTS": "*.result",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := ProcessRawExportMap(tC.input)

			assert.Equal(t, tC.expectedOutput, result)
		})
	}
}

func TestSimpleOutputExport_NoError(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "BITRISE_ARTIFACT_PATHS", "a/b.txt").Return(nil)

	exporter := OutputExporter{
		ExportValues:  "a/b.txt",
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
	}

	err := exporter.Export()
	assert.NoError(t, err)

	envRepository.AssertExpectations(t)
}

func TestSimpleOutputExport_Error(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "BITRISE_ARTIFACT_PATHS", "a/b.txt").Return(errors.New("some kind of error"))

	exporter := OutputExporter{
		ExportValues:  "a/b.txt",
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
	}

	err := exporter.Export()
	assert.EqualError(t, err, "failed to export pulled artifact locations, error: some kind of error")

	envRepository.AssertExpectations(t)
}

func TestPatternBasedOutputExport_SingleFile_NoError(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "TXT_FILES", "a.txt").Return(nil)
	envRepository.On("Set", "APK_FILES", "b.apk").Return(nil)

	exporter := OutputExporter{
		ExportValues:  "a.txt|b.apk",
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
		ExportPattern: map[string]string{
			"TXT_FILES": "*.txt",
			"APK_FILES": "*.apk",
		},
	}

	err := exporter.Export()
	assert.NoError(t, err)

	envRepository.AssertExpectations(t)
}

func TestPatternBasedOutputExport_MultipleFiles_NoError(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "TXT_FILES", "a.txt|b.txt").Return(nil)
	envRepository.On("Set", "APK_FILES", "b.apk").Return(nil)

	exporter := OutputExporter{
		ExportValues:  "a.txt|b.txt|b.apk|x.ipa",
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
		ExportPattern: map[string]string{
			"TXT_FILES": "*.txt",
			"APK_FILES": "*.apk",
		},
	}

	err := exporter.Export()
	assert.NoError(t, err)

	envRepository.AssertExpectations(t)
}

func TestPatternBasedOutputExport_MultipleFiles_MultipleExpressions_NoError(t *testing.T) {
	envRepository := new(mockenv.Repository)

	envRepository.On("Set", "TEXT_FILES", "a.txt|b.txt|d.docx").Return(nil)
	envRepository.On("Set", "APK_FILES", "b.apk").Return(nil)
	envRepository.On("Set", "ALL", "a.txt|b.txt|b.apk|x.ipa|d.docx").Return(nil)

	exporter := OutputExporter{
		ExportValues:  "a.txt|b.txt|b.apk|x.ipa|d.docx",
		Logger:        log.NewLogger(),
		EnvRepository: envRepository,
		ExportPattern: map[string]string{
			"TEXT_FILES": "*.txt,*.docx",
			"APK_FILES":  "*.apk",
			"ALL":        "*",
		},
	}

	err := exporter.Export()
	assert.NoError(t, err)

	envRepository.AssertExpectations(t)
}
