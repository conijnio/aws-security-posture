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
	"regexp"

	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"

	"os"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	json.Unmarshal(file, &event)
	return event
}

func readRawFindings(path string) ([]byte, []types.AwsSecurityFinding) {
	file, _ := os.ReadFile(path)

	var findings []types.AwsSecurityFinding
	json.Unmarshal(file, &findings)
	data, _ := json.Marshal(findings)
	return data, findings
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

func GetFindingsInput(report string) *securityhub.GetFindingsInput {
	return &securityhub.GetFindingsInput{
		MaxResults: int32(100),
		Filters:    GetFilters(report),
	}
}

func GetFindingsInputWithToken(report string, token string) *securityhub.GetFindingsInput {
	return &securityhub.GetFindingsInput{
		MaxResults: int32(100),
		NextToken:  aws.String(token),
		Filters:    GetFilters(report),
	}
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/collect-findings.json")
	expectedBody, rawFindings := readRawFindings("../../events/raw-findings.json")

	t.Run("Fetch 7 findings in 3 iterations", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInput("aws-foundational-security-best-practices"),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings[0:2], NextToken: aws.String("Page2")},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInputWithToken("aws-foundational-security-best-practices", "Page2"),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings[2:4], NextToken: aws.String("Page3")},
		})
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInputWithToken("aws-foundational-security-best-practices", "Page3"),
			Output:        &securityhub.GetFindingsOutput{Findings: rawFindings[4:7]},
		})

		stubber.Add(testtools.Stub{
			OperationName: "PutObject",
			Input:         &s3.PutObjectInput{Bucket: aws.String("my-sample-bucket"), Body: bytes.NewReader(expectedBody)},
			Output:        &s3.PutObjectOutput{},
			IgnoreFields:  []string{"Key"},
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

		regex, _ := regexp.Compile(fmt.Sprintf("%s/raw/[0-9]{4}/[0-9]{2}/[0-9]{2}/[0-9]{10}.json", response.Report))

		if regex.FindAllString(response.Key, -1) == nil {
			t.Errorf("Unexpected object key: %s", response.Key)
		}
	})

	t.Run("GetFindings raises error", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "GetFindings",
			Input:         GetFindingsInputWithToken("aws-foundational-security-best-practices", "Page3"),
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
			Input:         GetFindingsInput("aws-foundational-security-best-practices"),
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
