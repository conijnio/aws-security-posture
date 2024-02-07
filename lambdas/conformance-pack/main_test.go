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

		event := readEvent("../../events/conformance-pack.json")
		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, "Title", response.Accounts[0].GroupBy)
		assert.Equal(t, "acme-workload-development", response.Accounts[0].AccountName)
		assert.Equal(t, 4, len(response.Accounts[0].Controls))
	})

	t.Run("Invoke No AccountName in Mapping", func(t *testing.T) {
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
				},
			},
		})

		event := readEvent("../../events/conformance-pack.json")
		delete(event.AccountMapping, "111122223333")
		response, err := lambda.Handler(ctx, event)

		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, "Title", response.Accounts[0].GroupBy)
		assert.Equal(t, "", response.Accounts[0].AccountName)
		assert.Equal(t, 1, len(response.Accounts[0].Controls))
	})

	t.Run("Invoke additional accounts", func(t *testing.T) {
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
								ConfigRuleName: aws.String("lz-rule-2-aaaaa"),
							},
						},
					},
				},
			},
		})

		event := readEvent("../../events/conformance-pack.json")
		event.AccountMapping["111122224444"] = "acme-workload-testing"
		response, err := lambda.Handler(ctx, event)

		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response.Accounts))

		assert.Equal(t, "Title", response.Accounts[0].GroupBy)
		assert.Equal(t, "111122223333", response.Accounts[0].AccountId)
		assert.Equal(t, "acme-workload-development", response.Accounts[0].AccountName)
		assert.Equal(t, "my-sample-bucket", response.Accounts[0].Bucket)
		assert.Equal(t, "aws-foundational-security-best-practices/111122223333/2023/08/13/111111111111.json", response.Accounts[0].Key)
		assert.Equal(t, 2, len(response.Accounts[0].Controls))

		assert.Equal(t, "Title", response.Accounts[1].GroupBy)
		assert.Equal(t, "111122224444", response.Accounts[1].AccountId)
		assert.Equal(t, "acme-workload-testing", response.Accounts[1].AccountName)
		assert.Equal(t, "", response.Accounts[1].Bucket)
		assert.Equal(t, "", response.Accounts[1].Key)
		assert.Equal(t, 2, len(response.Accounts[1].Controls))

	})
}
