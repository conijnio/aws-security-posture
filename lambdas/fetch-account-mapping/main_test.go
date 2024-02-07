package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
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
	event := readEvent("../../events/fetch-account-mapping.json")

	t.Run("Invoke", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)

		stubber.Add(testtools.Stub{
			OperationName: "ListAccounts",
			Input: &organizations.ListAccountsInput{
				MaxResults: aws.Int32(20),
			},
			Output: &organizations.ListAccountsOutput{
				Accounts: []types.Account{
					{
						Id:   aws.String("111111111111"),
						Name: aws.String("acme-workload-build"),
					},
					{
						Id:   aws.String("111122223333"),
						Name: aws.String("acme-workload-development"),
					},
					{
						Id:   aws.String("111111111113"),
						Name: aws.String("acme-workload-test"),
					},
					{
						Id:   aws.String("111111111114"),
						Name: aws.String("acme-workload-acceptance"),
					},
					{
						Id:   aws.String("111111111115"),
						Name: aws.String("acme-workload-production"),
					},
				},
			},
		})

		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.ConformancePack, response.ConformancePack)
		assert.Equal(t, 1, len(response.Accounts))
		assert.Equal(t, 5, len(response.AccountMapping))
		assert.Equal(t, "acme-workload-build", response.AccountMapping["111111111111"])
		assert.Equal(t, "acme-workload-development", response.AccountMapping["111122223333"])
		assert.Equal(t, "acme-workload-test", response.AccountMapping["111111111113"])
		assert.Equal(t, "acme-workload-acceptance", response.AccountMapping["111111111114"])
		assert.Equal(t, "acme-workload-production", response.AccountMapping["111111111115"])
	})
}
