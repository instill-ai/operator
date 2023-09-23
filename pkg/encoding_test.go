package pkg

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	base64Op "github.com/instill-ai/operator/pkg/base64"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{
			Name:           "positive test case",
			Input:          "Hello, World!",
			ExpectedOutput: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			Name:           "already encoded string",
			Input:          "SGVsbG8sIFdvcmxkIQ==",
			ExpectedOutput: "SGVsbG8sIFdvcmxkIQ==",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.ExpectedOutput, base64Op.Encode(test.Input))
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		Name           string
		Input          string
		ExpectedOutput string
		ExpectedErr    error
	}{
		{
			Name:           "positive test case",
			Input:          "SGVsbG8sIFdvcmxkIQ==",
			ExpectedOutput: "Hello, World!",
		},
		{
			Name:           "negative test case",
			Input:          "Hello, World!",
			ExpectedOutput: "Hello, World!",
			ExpectedErr:    base64.CorruptInputError(5),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res, err := base64Op.Decode(test.Input)
			assert.Equal(t, test.ExpectedOutput, res)
			assert.Equal(t, test.ExpectedErr, err)
		})
	}
}
