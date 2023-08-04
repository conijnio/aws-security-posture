package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"os"
	"testing"
	"time"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	json.Unmarshal(file, &event)
	return event
}

func PutMetricDataInput(report string, accountId string, score float64) *cloudwatch.PutMetricDataInput {
	return &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("SecurityPosture"),
		MetricData: []types.MetricDatum{
			types.MetricDatum{
				Timestamp:  aws.Time(time.Unix(1691920532, 0)),
				MetricName: aws.String("Score"),
				Dimensions: []types.Dimension{
					types.Dimension{
						Name:  aws.String("Report"),
						Value: aws.String(report),
					},
					types.Dimension{
						Name:  aws.String("AccountId"),
						Value: aws.String(accountId),
					},
				},
				Value: aws.Float64(score),
			},
		},
	}
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/publish-metrics.json")

	t.Run("Publish Score", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)

		// TODO: Do 1 API Call for multiple metrics

		stubber.Add(testtools.Stub{
			OperationName: "PutMetricData",
			Input:         PutMetricDataInput("aws-foundational-security-best-practices", "111122223333", 80),
			Output:        &cloudwatch.PutMetricDataOutput{},
		})
		stubber.Add(testtools.Stub{
			OperationName: "PutMetricData",
			Input:         PutMetricDataInput("aws-foundational-security-best-practices", "333322221111", 90),
			Output:        &cloudwatch.PutMetricDataOutput{},
		})

		_, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)

		if err != nil {
			t.Errorf("Expected nil, but got %q", err)
		}
	})

	t.Run("Fail on PutMetricData", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "PutMetricData",
			Input:         PutMetricDataInput("aws-foundational-security-best-practices", "333322221111", 90),
			Error:         raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})
}
