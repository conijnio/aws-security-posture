package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"log"
	"os"
	"strings"
)

type Lambda struct {
	ctx    context.Context
	client *organizations.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.client = organizations.NewFromConfig(cfg)
	return m
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	response := Response{
		AccountId: request.AccountId,
		Bucket:    request.Bucket,
		Key:       request.Key,
	}
	x.ctx = ctx
	log.Printf("Fetching workload context for: %s", request.AccountId)

	account, err := x.client.DescribeAccount(x.ctx, &organizations.DescribeAccountInput{AccountId: aws.String(request.AccountId)})

	if err != nil {
		return response, err
	}

	accountName := x.platformOverwrite(request.AccountId, *account.Account.Name)
	workload, environment, err := x.resolveWorkloadAndEnvironment(accountName)
	response.Workload = workload
	response.Environment = environment

	return response, err
}

func (x *Lambda) platformOverwrite(accountId string, name string) string {
	accountString := os.Getenv("PLATFORM_ACCOUNTS")
	overwrites := strings.Split(accountString, ",")

	for _, overwriteAccountId := range overwrites {
		if accountId == overwriteAccountId {
			return fmt.Sprintf("%s-production", name)
		}
	}

	return name
}

func (x *Lambda) resolveWorkloadAndEnvironment(name string) (string, string, error) {
	workload := ""
	environment := "production"
	parts := strings.Split(name, "-")

	if len(parts) < 2 {
		return workload, environment, errors.New(fmt.Sprintf("Could not transform `%s` to a workload and environment name", name))
	}

	if len(parts) > 2 {
		workload = strings.Join(parts[1:len(parts)-1], "-")
		environment = parts[len(parts)-1]
	} else {
		workload = parts[len(parts)-1]
	}

	return workload, environment, nil
}
