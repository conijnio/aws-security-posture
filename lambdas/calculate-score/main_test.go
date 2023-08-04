package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"io"
	"os"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	json.Unmarshal(file, &event)
	return event
}

func readRawFindings(path string) []byte {
	file, _ := os.ReadFile(path)

	var findings []types.AwsSecurityFinding
	_ = json.Unmarshal(file, &findings)
	data, _ := json.Marshal(findings[0:4])
	return data
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/calculate-score.json")
	source := readRawFindings("../../events/raw-findings.json")

	t.Run("Calculate score", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/111122223333/2023/08/13/111111111111.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(source))},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)

		if err != nil {
			t.Errorf("Expected nil, but got %q", err)
		}

		if response.AccountId != event.AccountId {
			t.Errorf("Expected %s, but got %s", event.AccountId, response.AccountId)
		}

		if response.Score != 50 {
			t.Errorf("Expected 50, but got %f", response.Score)
		}

	})

	t.Run("Fail on download", func(t *testing.T) {
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
}
