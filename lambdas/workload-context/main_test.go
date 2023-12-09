package main

import (
	"context"
	"encoding/json"
	"errors"
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
	json.Unmarshal(file, &event)
	return event
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	event := readEvent("../../events/workload-context.json")

	t.Run("Fail on DescribeAccount", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)
		raiseErr := &testtools.StubError{Err: errors.New("failed")}
		stubber.Add(testtools.Stub{
			OperationName: "DescribeAccount",
			Input: &organizations.DescribeAccountInput{
				AccountId: aws.String(event.AccountId),
			},
			Error: raiseErr,
		})

		_, err := lambda.Handler(ctx, event)
		testtools.VerifyError(err, raiseErr, t)
		testtools.ExitTest(stubber, t)
	})

	t.Run("Get Workload Context", func(t *testing.T) {
		stubber := testtools.NewStubber()

		stubber.Add(testtools.Stub{
			OperationName: "DescribeAccount",
			Input: &organizations.DescribeAccountInput{
				AccountId: aws.String(event.AccountId),
			},
			Output: &organizations.DescribeAccountOutput{
				Account: &types.Account{Name: aws.String("prefix-my-workload-development")},
			},
		})

		lambda := New(*stubber.SdkConfig)
		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Bucket, response.Bucket)
		assert.Equal(t, event.Key, response.Key)
		assert.Equal(t, event.AccountId, response.AccountId)
		assert.Equal(t, "my-workload", response.Workload)
		assert.Equal(t, "development", response.Environment)
	})

	t.Run("Resolve Workload and Environment", func(t *testing.T) {
		stubber := testtools.NewStubber()
		lambda := New(*stubber.SdkConfig)

		workload, environment, err := lambda.resolveWorkloadAndEnvironment("prefix-workload-development")
		assert.NoError(t, err)
		assert.Equal(t, "workload", workload)
		assert.Equal(t, "development", environment)

		workload, environment, err = lambda.resolveWorkloadAndEnvironment("prefix-my-workload-testing")
		assert.NoError(t, err)
		assert.Equal(t, "my-workload", workload)
		assert.Equal(t, "testing", environment)

		workload, environment, err = lambda.resolveWorkloadAndEnvironment("prefix-my-other-workload-acceptance")
		assert.NoError(t, err)
		assert.Equal(t, "my-other-workload", workload)
		assert.Equal(t, "acceptance", environment)

		workload, environment, err = lambda.resolveWorkloadAndEnvironment("prefix")
		assert.Error(t, err, "we always expect 3 parts a <prefix>-<workload>-<environment>")

		// Account names without an environment are considered production
		workload, environment, err = lambda.resolveWorkloadAndEnvironment("prefix-workload")
		assert.NoError(t, err)
		assert.Equal(t, "workload", workload)
		assert.Equal(t, "production", environment)

		testtools.ExitTest(stubber, t)
	})

	t.Run("Get Workload Context from log-archive PlatformAccounts", func(t *testing.T) {
		_ = os.Setenv("PLATFORM_ACCOUNTS", "666655554444,333322221111")
		logPrefixEvent := readEvent("../../events/workload-context.json")
		logPrefixEvent.AccountId = "666655554444"
		stubber := testtools.NewStubber()

		stubber.Add(testtools.Stub{
			OperationName: "DescribeAccount",
			Input: &organizations.DescribeAccountInput{
				AccountId: aws.String(logPrefixEvent.AccountId),
			},
			Output: &organizations.DescribeAccountOutput{
				Account: &types.Account{Name: aws.String("prefix-log-archive")},
			},
		})

		lambda := New(*stubber.SdkConfig)
		response, err := lambda.Handler(ctx, logPrefixEvent)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, logPrefixEvent.Bucket, response.Bucket)
		assert.Equal(t, logPrefixEvent.Key, response.Key)
		assert.Equal(t, logPrefixEvent.AccountId, response.AccountId)
		assert.Equal(t, "log-archive", response.Workload)
		assert.Equal(t, "production", response.Environment)

		_ = os.Setenv("PLATFORM_ACCOUNTS", "")
	})

	t.Run("Get Workload Context from shared-services PlatformAccounts", func(t *testing.T) {
		_ = os.Setenv("PLATFORM_ACCOUNTS", "666655554444,333322221111")
		sharedServicesEvent := readEvent("../../events/workload-context.json")
		sharedServicesEvent.AccountId = "333322221111"
		stubber := testtools.NewStubber()

		stubber.Add(testtools.Stub{
			OperationName: "DescribeAccount",
			Input: &organizations.DescribeAccountInput{
				AccountId: aws.String(sharedServicesEvent.AccountId),
			},
			Output: &organizations.DescribeAccountOutput{
				Account: &types.Account{Name: aws.String("prefix-shared-services")},
			},
		})

		lambda := New(*stubber.SdkConfig)
		response, err := lambda.Handler(ctx, sharedServicesEvent)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, sharedServicesEvent.Bucket, response.Bucket)
		assert.Equal(t, sharedServicesEvent.Key, response.Key)
		assert.Equal(t, sharedServicesEvent.AccountId, response.AccountId)
		assert.Equal(t, "shared-services", response.Workload)
		assert.Equal(t, "production", response.Environment)

		_ = os.Setenv("PLATFORM_ACCOUNTS", "")
	})

	t.Run("Get Workload Context from workload with PLATFORM_ACCOUNTS config", func(t *testing.T) {
		_ = os.Setenv("PLATFORM_ACCOUNTS", "666655554444,333322221111")
		stubber := testtools.NewStubber()

		stubber.Add(testtools.Stub{
			OperationName: "DescribeAccount",
			Input: &organizations.DescribeAccountInput{
				AccountId: aws.String(event.AccountId),
			},
			Output: &organizations.DescribeAccountOutput{
				Account: &types.Account{Name: aws.String("prefix-my-workload-development")},
			},
		})

		lambda := New(*stubber.SdkConfig)
		response, err := lambda.Handler(ctx, event)
		testtools.ExitTest(stubber, t)
		assert.NoError(t, err)
		assert.Equal(t, event.Bucket, response.Bucket)
		assert.Equal(t, event.Key, response.Key)
		assert.Equal(t, event.AccountId, response.AccountId)
		assert.Equal(t, "my-workload", response.Workload)
		assert.Equal(t, "development", response.Environment)

		_ = os.Setenv("PLATFORM_ACCOUNTS", "")
	})
}
