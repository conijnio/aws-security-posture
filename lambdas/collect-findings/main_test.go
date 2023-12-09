package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"github.com/stretchr/testify/assert"
	"regexp"

	"os"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func readRawFindings(path string) []types.AwsSecurityFinding {
	file, _ := os.ReadFile(path)

	var findings []types.AwsSecurityFinding
	_ = json.Unmarshal(file, &findings)

	return findings
}

func readStrippedFindings(path string) []byte {
	file, _ := os.ReadFile(path)
	var findings []*Finding
	_ = json.Unmarshal(file, &findings)
	data, _ := json.Marshal(findings)

	return data
}

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

func GetFindingsInput(report string, results int, token string) *securityhub.GetFindingsInput {
	var awsToken *string

	if token != "" {
		awsToken = aws.String(token)
	}

	return &securityhub.GetFindingsInput{
		MaxResults: aws.Int32(int32(results)),
		Filters:    GetFilters(report),
		NextToken:  awsToken,
	}
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/collect-findings.json")
	rawFindings := readRawFindings("../../events/raw-findings.json")
	strippedFindings := readStrippedFindings("../../events/stripped-findings.json")

	t.Run("Fetch 7 findings in 1 iterations", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 100, ""),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket"), Body: bytes.NewReader(strippedFindings)},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key"},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)

		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)
		regex, _ := regexp.Compile(fmt.Sprintf("%s/raw/[0-9]{4}/[0-9]{2}/[0-9]{2}/[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}.json", response.Report))

		if regex.FindAllString(response.Findings[0], -1) == nil {
			t.Errorf("Unexpected object key: %s", response.Findings[0])
		}
	})

	t.Run("GetFindings raises error", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 100, ""),
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})

	t.Run("PutObject raises error", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}

		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 100, ""),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings},
		})
		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket")},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key", "Body"},
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})
}

func TestLambdaFunctionLoop(t *testing.T) {
	_ = os.Setenv("MAX_RESULTS", "3")

	ctx := context.Background()
	event := readEvent("../../events/collect-findings.json")
	rawFindings := readRawFindings("../../events/raw-findings.json")

	t.Run("Exhaust Lambda and trigger the StepFunction loop", func(t *testing.T) {

		// Simulate 1st invocation
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 3, ""),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings[0:3], NextToken: aws.String("Page2")},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket")},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key", "Body"},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)

		// Simulate 2nd invocation
		stubber = testtools.NewStubber()
		lambda = New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 3, "Page2"),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings[3:6], NextToken: aws.String("Page3")},
		})
		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket")},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key", "Body"},
		})

		//event.Filter = nil
		event.NextToken = response.NextToken
		event.Findings = response.Findings
		event.Timestamp = response.Timestamp

		response, err = lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)

		// Simulate 3rd invocation
		stubber = testtools.NewStubber()
		lambda = New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 3, "Page3"),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings[6:7]},
		})
		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket")},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key", "Body"},
		})

		event.NextToken = response.NextToken
		event.Findings = response.Findings
		event.Timestamp = response.Timestamp

		response, err = lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)

	})

	t.Run("GetFindings raises error with token", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 100, "page1"),
			Error:         raiseErr,
		})

		event.NextToken = "page1"
		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})

	_ = os.Setenv("MAX_RESULTS", "")
	_ = os.Setenv("MAX_PAGES", "")
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

func TestAggregationPassing(t *testing.T) {
	rawFindings := readRawFindings("../../events/raw-findings.json")

	t.Run("Empty aggregation stays empty", func(t *testing.T) {
		ctx := context.Background()

		event := Request{
			Bucket:       "my-sample-bucket",
			Report:       "aws-foundational-security-best-practices",
			Filter:       *GetFilters("aws-foundational-security-best-practices"),
			FindingCount: 2,
			Findings: []string{
				"my/first/batch.json",
				"my/second/batch.json",
			},
		}

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 100, ""),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket")},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key", "Body"},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)
		assert.Equal(t, 0, len(response.AggregatedFindings))
	})

	t.Run("Populated aggregation stays the same", func(t *testing.T) {
		ctx := context.Background()

		event := Request{
			Bucket:       "my-sample-bucket",
			Report:       "aws-foundational-security-best-practices",
			Filter:       *GetFilters("aws-foundational-security-best-practices"),
			FindingCount: 2,
			Findings: []string{
				"my/first/batch.json",
				"my/second/batch.json",
			},
			AggregatedFindings: []string{
				"my/first/aggregated/batch.json",
				"my/second/aggregated/batch.json",
			},
		}

		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices", 100, ""),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket")},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key", "Body"},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Report, response.Report)
		assert.Equal(t, event.Bucket, response.Bucket)
		assert.Equal(t, 3, response.FindingCount)
		assert.Equal(t, 3, len(response.Findings))
		assert.Equal(t, 2, len(response.AggregatedFindings))
	})

}
