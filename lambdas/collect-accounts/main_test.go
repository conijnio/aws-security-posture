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

func readRawFindings(path string) ([]byte, []byte, []byte) {
	file, _ := os.ReadFile(path)

	var findings []types.AwsSecurityFinding
	_ = json.Unmarshal(file, &findings)
	data, _ := json.Marshal(findings)
	dataset1, _ := json.Marshal(findings[0:4])
	dataset2, _ := json.Marshal(findings[4:7])
	return data, dataset1, dataset2
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/collect-accounts.json")
	source, dataset1, dataset2 := readRawFindings("../../events/raw-findings.json")

	t.Run("Read raw findings and split based on AccountId", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/111111111111.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(source))},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input: &s3.PutObjectInput{
				Bucket: aws.String("my-sample-bucket"),
				Key:    aws.String("aws-foundational-security-best-practices/111122223333/2023/08/13/1691920532.json"),
				Body:   bytes.NewReader(dataset1),
			},
			Output: &s3.PutObjectOutput{},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input: &s3.PutObjectInput{
				Bucket: aws.String("my-sample-bucket"),
				Key:    aws.String("aws-foundational-security-best-practices/333322221111/2023/08/13/1691920532.json"),
				Body:   bytes.NewReader(dataset2),
			},
			Output:       &s3.PutObjectOutput{},
			IgnoreFields: []string{"Key"},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)

		if err != nil {
			t.Errorf("Expected nil, but got %q", err)
		}

		if response.Report != event.Report {
			t.Errorf("Expected %s, but got %s", event.Report, response.Report)
		}

		if response.Bucket != event.Bucket {
			t.Errorf("Expected %s, but got %s", event.Bucket, response.Bucket)
		}

		for _, account := range response.Accounts {
			if account.AccountId != "111122223333" && account.AccountId != "333322221111" {
				t.Errorf("Expected AccountId to be 111122223333 or 333322221111")
			}
		}

	})

	t.Run("Fail on download", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/111111111111.json")},
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})

	t.Run("Fail on upload", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/111111111111.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(source))},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input: &s3.PutObjectInput{
				Bucket: aws.String("my-sample-bucket"),
				Key:    aws.String("aws-foundational-security-best-practices/111122223333/2023/08/13/1691920532.json"),
				Body:   bytes.NewReader(dataset1),
			},
			Error: raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})
}
