package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
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
	x.ctx = ctx
	response := Response{
		Report:          request.Report,
		Bucket:          request.Bucket,
		Timestamp:       request.Timestamp,
		ConformancePack: request.ConformancePack,
		Accounts:        request.Accounts,
		AccountMapping:  map[string]string{},
	}

	paginator := organizations.NewListAccountsPaginator(x.client, &organizations.ListAccountsInput{
		MaxResults: aws.Int32(20),
	})

	pageNum := 0
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return response, err
		}
		for _, account := range output.Accounts {
			response.AccountMapping[*account.Id] = *account.Name
		}
		pageNum++
	}

	return response, nil
}
