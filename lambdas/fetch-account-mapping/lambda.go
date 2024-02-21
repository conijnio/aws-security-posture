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
		Report:    request.Report,
		Bucket:    request.Bucket,
		Timestamp: request.Timestamp,
		Accounts:  []Account{},
	}

	controls := request.Accounts[0].Controls
	mapping, err := x.resolveMapping()

	if err != nil {
		return response, err
	}

	for accountId, accountName := range mapping {
		found := false
		for _, account := range request.Accounts {
			if account.AccountId == accountId {
				if account.AccountName == "" {
					account.AccountName = accountName
				}
				found = true
				response.Accounts = append(response.Accounts, account)
				break
			}
		}

		if !found {
			response.Accounts = append(response.Accounts, Account{
				AccountId:   accountId,
				AccountName: accountName,
				Controls:    controls,
			})
		}
	}

	return response, nil
}

func (x *Lambda) resolveMapping() (map[string]string, error) {
	paginator := organizations.NewListAccountsPaginator(x.client, &organizations.ListAccountsInput{
		MaxResults: aws.Int32(20),
	})

	mapping := map[string]string{}
	pageNum := 0
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return mapping, err
		}
		for _, account := range output.Accounts {
			mapping[*account.Id] = *account.Name
		}
		pageNum++
	}

	return mapping, nil
}
