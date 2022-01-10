package export

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
*.apk - DOWNLOADED_APKS
*.result - DOWNLOADED_TEST_RESULTS
`,
			expectedOutput: map[string]string{
				"DOWNLOADED_APKS":         "*.apk",
				"DOWNLOADED_TEST_RESULTS": "*.result",
			},
		},
		{
			desc: "when multiple expressions presents in the expression field",
			input: `
*.apk,*ipa - ARTIFACTS
*.txt,*docx - TEXTS
`,
			expectedOutput: map[string]string{
				"ARTIFACTS":         "*.apk,*ipa",
				"TEXTS": "*.txt,*docx",
			},
		},
		{
			desc: "when the input has a messed up line",
			input: `
*.apk - DOWNLOADED_APKS
*.ipa -
*.result - DOWNLOADED_TEST_RESULTS
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
