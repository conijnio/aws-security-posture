package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func GetFilters(report string) *types.AwsSecurityFindingFilters {
	return &types.AwsSecurityFindingFilters{
		GeneratorId: []types.StringFilter{
			{
				Comparison: "PREFIX",
				Value:      aws.String(report),
			},
		},
		RecordState: []types.StringFilter{
			{
				Comparison: "NOT_EQUALS",
				Value:      aws.String("ARCHIVED"),
			},
		},
		WorkflowStatus: []types.StringFilter{
			{
				Comparison: "NOT_EQUALS",
				Value:      aws.String("SUPPRESSED"),
			},
		},
	}
}

func generateFinding(prefix string, index int) Finding {
	return Finding{
		Id: fmt.Sprintf("%s-%d", prefix, index),
	}
}

func generateFindings(prefix string, count int) []Finding {
	var findings []Finding
	for i := 0; i < count; i++ {
		findings = append(findings, generateFinding(prefix, i))
	}
	return findings
}

func toReader(findings []Finding) *bytes.Reader {
	data, _ := json.Marshal(findings)
	return bytes.NewReader(data)
}

func toReadCloser(findings []Finding) io.ReadCloser {
	return io.NopCloser(toReader(findings))
}

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func TestHandler(t *testing.T) {
	event := readEvent("../../events/aggregate-findings.json")

	t.Run("Aggregate 2 Findings", func(t *testing.T) {

		ctx := context.Background()
		firstBatch := generateFindings("first", 10)
		secondBatch := generateFindings("second", 10)
		expectedBatch := append(firstBatch, secondBatch...)

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("my/first/batch.json")},
			Output:        &s3.GetObjectOutput{Body: toReadCloser(firstBatch)},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("my/second/batch.json")},
			Output:        &s3.GetObjectOutput{Body: toReadCloser(secondBatch)},
		})
		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket"), Body: toReader(expectedBatch)},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key"},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)
		assert.Equal(t, event.Controls, response.Controls)
		assert.Equal(t, event.GroupBy, response.GroupBy)
		assert.Equal(t, 0, response.FindingCount)
		assert.Equal(t, 0, len(response.Findings))
		assert.Equal(t, 1, len(response.AggregatedFindings))
	})

	t.Run("Fail on downloading finding files", func(t *testing.T) {

		ctx := context.Background()
		firstBatch := generateFindings("first", 10)

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("my/first/batch.json")},
			Output:        &s3.GetObjectOutput{Body: toReadCloser(firstBatch)},
		})
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("my/second/batch.json")},
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.Error(t, err)
	})

	t.Run("Fail on uploading aggregated findings", func(t *testing.T) {

		ctx := context.Background()
		firstBatch := generateFindings("first", 10)
		secondBatch := generateFindings("second", 10)
		expectedBatch := append(firstBatch, secondBatch...)

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("my/first/batch.json")},
			Output:        &s3.GetObjectOutput{Body: toReadCloser(firstBatch)},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetObject",
			Input:         &s3.GetObjectInput{Bucket: aws.String("my-sample-bucket"), Key: aws.String("my/second/batch.json")},
			Output:        &s3.GetObjectOutput{Body: toReadCloser(secondBatch)},
		})
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket"), Body: toReader(expectedBatch)},
			IgnoreFields:  []string{"Key"},
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.Error(t, err)
	})
}
