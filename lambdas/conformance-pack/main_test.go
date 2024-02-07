package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/awsdocs/aws-doc-sdk-examples/gov2/testtools"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func readEvent(path string) Request {
	file, _ := os.ReadFile(path)

	var event Request
	_ = json.Unmarshal(file, &event)
	return event
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/conformance-pack.json")

	t.Run("Invoke", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)

		stubber.Add(testtools.Stub{
			OperationName: "GetConformancePackComplianceDetails",
			Input: &configservice.GetConformancePackComplianceDetailsInput{
				ConformancePackName: aws.String("OrgConformsPack-lz-framework-xxxxxxxx"),
				Limit:               100,
			},
			Output: &configservice.GetConformancePackComplianceDetailsOutput{
				ConformancePackName: aws.String("OrgConformsPack-lz-framework-xxxxxxxx"),
				ConformancePackRuleEvaluationResults: []types.ConformancePackEvaluationResult{
					{
						EvaluationResultIdentifier: &types.EvaluationResultIdentifier{
							EvaluationResultQualifier: &types.EvaluationResultQualifier{
								ConfigRuleName: aws.String("lz-rule-1-aaaaa"),
							},
						},
					},
					{
						EvaluationResultIdentifier: &types.EvaluationResultIdentifier{
							EvaluationResultQualifier: &types.EvaluationResultQualifier{
								ConfigRuleName: aws.String("lz-rule-1-bbbbb"),
							},
						},
					},
					{
						EvaluationResultIdentifier: &types.EvaluationResultIdentifier{
							EvaluationResultQualifier: &types.EvaluationResultQualifier{
								ConfigRuleName: aws.String("lz-rule-2-aaaaa"),
							},
						},
					},
					{
						EvaluationResultIdentifier: &types.EvaluationResultIdentifier{
							EvaluationResultQualifier: &types.EvaluationResultQualifier{
								ConfigRuleName: aws.String("lz-rule-2-bbbbbb"),
							},
						},
					},
					{
						EvaluationResultIdentifier: &types.EvaluationResultIdentifier{
							EvaluationResultQualifier: &types.EvaluationResultQualifier{
								ConfigRuleName: aws.String("lz-rule-3-cccccc"),
							},
						},
					},
					{
						EvaluationResultIdentifier: &types.EvaluationResultIdentifier{
							EvaluationResultQualifier: &types.EvaluationResultQualifier{
								ConfigRuleName: aws.String("lz-rule-4-ddddd"),
							},
						},
					},
				},
			},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, "Title", response.Accounts[0].GroupBy)
		assert.Equal(t, 4, len(response.Accounts[0].Controls))
	})
}
