package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func toReader(findings []string) *bytes.Reader {
	data, _ := json.Marshal(findings)
	return bytes.NewReader(data)
}

func TestHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("Invoke", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input: &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket"), Body: toReader([]string{
				"lz-rule-1", "lz-rule-2", "lz-rule-3", "lz-rule-4",
			})},
			Output:       &s3.PutObjectOutput{},
			IgnoreFields: []string{"Key"},
		})

		event := readEvent("../../events/custom-rules.json")
		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, "Title", response.GroupBy)
		assert.Equal(t, true, strings.HasPrefix(response.Controls, "my-security-framework/controls/"))
	})
}
