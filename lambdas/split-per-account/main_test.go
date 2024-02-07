package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func readStrippedFindings(path string) ([]Finding, []byte, []byte) {
	file, _ := os.ReadFile(path)

	var findings []Finding
	_ = json.Unmarshal(file, &findings)
	dataset1, _ := json.Marshal(findings[0:4])
	dataset2, _ := json.Marshal(findings[4:7])
	return findings, dataset1, dataset2
}

func getPages(findings []Finding) ([]byte, []byte, []byte) {
	page1, _ := json.Marshal(findings[0:3])
	page2, _ := json.Marshal(findings[3:6])
	page3, _ := json.Marshal(findings[6:7])

	return page1, page2, page3
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	findings, dataset1, dataset2 := readStrippedFindings("../../events/stripped-findings.json")
	page1, page2, page3 := getPages(findings)

	t.Run("Read raw findings and split based on AccountId", func(t *testing.T) {

		event := Request{
			Bucket:       "my-sample-bucket",
			Report:       "aws-foundational-security-best-practices",
			Timestamp:    1691920532,
			FindingCount: 3,
			Findings: []string{
				"aws-foundational-security-best-practices/raw/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json",
				"aws-foundational-security-best-practices/raw/2023/08/13/e43e3ef4-9380-11ee-b9d1-0242ac120002.json",
				"aws-foundational-security-best-practices/raw/2023/08/13/e88c8e3e-9380-11ee-b9d1-0242ac120002.json",
			},
		}

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page1))},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/e43e3ef4-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page2))},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/e88c8e3e-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page3))},
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

		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)

		for _, account := range response.Accounts {
			if account.AccountId != "111122223333" && account.AccountId != "333322221111" {
				t.Errorf("Expected AccountId to be 111122223333 or 333322221111")
			}

			if account.AccountName != "acme-workload-development" && account.AccountName != "acme-workload-test" {
				t.Errorf("Expected AccountName to be acme-workload-development or acme-workload-test")
			}
		}

	})

	t.Run("Read 2 raw findings and 1 aggregated and split based on AccountId", func(t *testing.T) {
		event := Request{
			Bucket:       "my-sample-bucket",
			Report:       "aws-foundational-security-best-practices",
			Timestamp:    1691920532,
			FindingCount: 2,
			Findings: []string{
				"aws-foundational-security-best-practices/raw/2023/08/13/e43e3ef4-9380-11ee-b9d1-0242ac120002.json",
				"aws-foundational-security-best-practices/raw/2023/08/13/e88c8e3e-9380-11ee-b9d1-0242ac120002.json",
			},
			AggregatedFindings: []string{
				"aws-foundational-security-best-practices/aggregated/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json",
			},
		}

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/aggregated/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page1))},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/e43e3ef4-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page2))},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/e88c8e3e-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page3))},
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

		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)

		for _, account := range response.Accounts {
			if account.AccountId != "111122223333" && account.AccountId != "333322221111" {
				t.Errorf("Expected AccountId to be 111122223333 or 333322221111")
			}
		}

	})

	t.Run("Fail on download", func(t *testing.T) {
		event := readEvent("../../events/split-per-account.json")
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
		event := readEvent("../../events/split-per-account.json")
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/dfcec91a-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page1))},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/e43e3ef4-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page2))},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("aws-foundational-security-best-practices/raw/2023/08/13/e88c8e3e-9380-11ee-b9d1-0242ac120002.json")},
			Output:        &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(page3))},
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
		testtools.ExitTest(stubber, t)
		testtools.VerifyError(err, raiseErr, t)
	})
}
