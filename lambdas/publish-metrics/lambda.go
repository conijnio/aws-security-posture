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
		var data []types.MetricDatum

		data = append(data, types.MetricDatum{
			Timestamp:  aws.Time(time.Unix(request.Timestamp, 0)),
			MetricName: aws.String("Score"),
			Dimensions: x.renderDimensions(request.Report, calculatedScore.Workload, calculatedScore.Environment),
			Value:      aws.Float64(calculatedScore.Score),
			Unit:       types.StandardUnitPercent,
		})

		data = append(data, types.MetricDatum{
			Timestamp:  aws.Time(time.Unix(request.Timestamp, 0)),
			MetricName: aws.String("Controls"),
			Dimensions: x.renderDimensions(request.Report, calculatedScore.Workload, calculatedScore.Environment),
			Value:      aws.Float64(float64(calculatedScore.ControlCount)),
			Unit:       types.StandardUnitCount,
		})

		data = append(data, types.MetricDatum{
			Timestamp:  aws.Time(time.Unix(request.Timestamp, 0)),
			MetricName: aws.String("Findings"),
			Dimensions: x.renderDimensions(request.Report, calculatedScore.Workload, calculatedScore.Environment),
			Value:      aws.Float64(float64(calculatedScore.FindingCount)),
			Unit:       types.StandardUnitCount,
		})

		// NOTE: The maximum number of metrics is 1000, we can optimize the API usage in the future here.
		err := x.publishBatch(data)

		if err != nil {
			return Response{}, err
		}
	}

	return Response{}, nil
}

func (x *Lambda) publishBatch(data []types.MetricDatum) error {
	_, err := x.client.PutMetricData(x.ctx, &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String("SecurityPosture"),
		MetricData: data,
	})

	return err
}

func (x *Lambda) renderDimensions(report string, workload string, environment string) []types.Dimension {
	return []types.Dimension{
		{
			Name:  aws.String("Report"),
			Value: aws.String(report),
		},
		{
			Name:  aws.String("Workload"),
			Value: aws.String(workload),
		},
		{
			Name:  aws.String("Environment"),
			Value: aws.String(environment),
		},
	}
}
