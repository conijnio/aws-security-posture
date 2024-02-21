package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
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
			OperationName: "DescribeStandardsControls",
			Input: &securityhub.DescribeStandardsControlsInput{
				StandardsSubscriptionArn: aws.String("arn:aws:securityhub:eu-central-1:000000000000:subscription/cis-aws-foundations-benchmark/v/1.2.0"),
				MaxResults:               aws.Int32(100),
			},
			Output: &securityhub.DescribeStandardsControlsOutput{
				Controls: []types.StandardsControl{
					{StandardsControlArn: aws.String("C1"), ControlStatus: types.ControlStatusEnabled},
					{StandardsControlArn: aws.String("C2"), ControlStatus: types.ControlStatusEnabled},
					{StandardsControlArn: aws.String("C3"), ControlStatus: types.ControlStatusEnabled},
					{StandardsControlArn: aws.String("C4"), ControlStatus: types.ControlStatusEnabled},
					{StandardsControlArn: aws.String("C4"), ControlStatus: types.ControlStatusDisabled},
					{StandardsControlArn: aws.String("C4"), ControlStatus: types.ControlStatusDisabled},
				},
			},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket"), Body: toReader([]string{"C1", "C2", "C3", "C4"})},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key"},
		})

		event := readEvent("../../events/subscription.json")
		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, "GeneratorId", response.GroupBy)
		assert.Equal(t, true, strings.HasPrefix(response.Controls, "aws-foundational-security-best-practices/controls/"))
	})
}
