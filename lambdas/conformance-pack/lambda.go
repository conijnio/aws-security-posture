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
	response := Response{
		Report:  request.Report,
		Bucket:  request.Bucket,
		GroupBy: "Title",
		Filter:  request.Filter,
		Controls: []string{
			// TODO: Let's query this from the API
			"lz-s3-access-logging-conformance-pack",
		},
	}
	x.ctx = ctx

	log.Printf("Loading Conformance Pack Context: %s", request.ConformancePack)

	return response, nil
}
