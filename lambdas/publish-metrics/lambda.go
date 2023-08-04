package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"time"
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

	for _, calculatedScore := range request.Accounts {
		_, err := x.client.PutMetricData(x.ctx, &cloudwatch.PutMetricDataInput{
			Namespace: aws.String("SecurityPosture"),
			MetricData: []types.MetricDatum{
				types.MetricDatum{
					Timestamp:  aws.Time(time.Unix(request.Timestamp, 0)),
					MetricName: aws.String("Score"),
					Dimensions: []types.Dimension{
						types.Dimension{
							Name:  aws.String("Report"),
							Value: aws.String(request.Report),
						},
						types.Dimension{
							Name:  aws.String("AccountId"),
							Value: aws.String(calculatedScore.AccountId),
						},
					},
					Value: aws.Float64(calculatedScore.Score),
				},
			},
		})
		if err != nil {
			return Response{}, err
		}
	}

	return Response{}, nil
}
