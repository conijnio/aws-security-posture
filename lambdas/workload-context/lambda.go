package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"os"
	"strings"
)

type Lambda struct {
	ctx context.Context
}

func New(cfg aws.Config) *Lambda {
	return new(Lambda)
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	response := Response{
		AccountId:   request.AccountId,
		AccountName: request.AccountName,
		Bucket:      request.Bucket,
		Key:         request.Key,
		GroupBy:     request.GroupBy,
		Controls:    request.Controls,
	}
	x.ctx = ctx

	name := x.platformOverwrite(request.AccountId, request.AccountName)
	workload, environment, err := x.resolveWorkloadAndEnvironment(name)
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
