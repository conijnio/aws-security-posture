package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func readRawFindings(path string) []*Finding {
	file, _ := os.ReadFile(path)

	var findings []*Finding
	_ = json.Unmarshal(file, &findings)

	return findings
}

func streamFindingData(findings []*Finding) io.ReadCloser {
	data, _ := json.Marshal(findings)
	return io.NopCloser(bytes.NewReader(data))
}

func streamControls(controls []string) io.ReadCloser {
	data, _ := json.Marshal(controls)
	return io.NopCloser(bytes.NewReader(data))
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/calculate-score.json")
	source := readRawFindings("../../events/stripped-findings.json")

	t.Run("Calculate score", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/111122223333/2023/08/13/111111111111.json")},
			Output:        &s3.GetObjectOutput{Body: streamFindingData(source[0:4])},
		})

		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/controls/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json")},
			Output: &s3.GetObjectOutput{Body: streamControls([]string{
				"arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0/rule/4.3",
				"arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0/rule/4.4",
			})},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)

		assert.NoError(t, err)
		assert.Equal(t, event.AccountId, response.AccountId)
		assert.Equal(t, float64(50), response.Score)
		assert.Equal(t, 2, response.ControlCount)
		assert.Equal(t, 4, response.FindingCount)
		assert.Equal(t, event.Workload, response.Workload)
		assert.Equal(t, event.Environment, response.Environment)
	})

	t.Run("Fail on findings download", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/111122223333/2023/08/13/111111111111.json")},
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})

	t.Run("Fail on control download", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/111122223333/2023/08/13/111111111111.json")},
			Output:        &s3.GetObjectOutput{Body: streamFindingData(source[0:4])},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/controls/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json")},
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})

	t.Run("No Bucket or Key", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)

		eventModified := event
		eventModified.Bucket = ""
		eventModified.Key = ""

		response, err := lambda.Handler(ctx, eventModified)

		assert.NoError(t, err)
		assert.Equal(t, 0, int(response.Score))
		assert.Equal(t, 0, response.ControlPassedCount)
		assert.Equal(t, 0, response.ControlFailedCount)
		assert.Equal(t, 0, response.ControlCount)
		testtools.ExitTest(stubber, t)
	})
}
