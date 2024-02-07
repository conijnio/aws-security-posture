package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"log"
)

type Lambda struct {
	ctx    context.Context
	client *cloudwatch.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.client = cloudwatch.NewFromConfig(cfg)
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

	log.Printf("Loading Conformance Pack Context: %s", request.ConformancePack)
	groupBy := "Title"
	controls := []string{
		// TODO: Let's query this from the API
		"lz-s3-access-logging-conformance-pack",
	}

	for _, account := range request.Accounts {
		response.Accounts = append(response.Accounts, Account{
			AccountId: account.AccountId,
			Bucket:    account.Bucket,
			Key:       account.Key,
			GroupBy:   groupBy,
			Controls:  controls,
		})
	}

	return response, nil
}
