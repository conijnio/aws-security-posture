package main

import (
	"context"
	"encoding/json"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/conformance-pack.json")

	t.Run("Invoke", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)

		assert.Equal(t, "Title", response.GroupBy)
		assert.Equal(t, 1, len(response.Controls))
	})
}
