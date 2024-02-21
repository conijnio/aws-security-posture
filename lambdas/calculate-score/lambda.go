package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
)

type Lambda struct {
	ctx      context.Context
	s3Client *s3.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.s3Client = s3.NewFromConfig(cfg)
	return m
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	response := Response{
		AccountId:   request.AccountId,
		AccountName: request.AccountName,
		Workload:    request.Workload,
		Environment: request.Environment,
		Score:       0,
	}

	if request.Bucket == "" || request.Key == "" || request.Controls == "" {
		return response, nil
	}

	x.ctx = ctx
	log.Printf("Calculating the security score for: %s", request.AccountId)

	findings, err := x.downloadFindings(request.Bucket, request.Key)

	if err != nil {
		return response, err
	}

	controls, err := x.downloadControls(request.Bucket, request.Controls)
	if err != nil {
		return response, err
	}

	calc := NewCalculator(controls)

	for _, finding := range findings {
		calc.ProcessFinding(finding, request.GroupBy)
	}

	response.Score = calc.Score()
	response.ControlCount = calc.ControlCount()
	response.ControlFailedCount = calc.ControlFailedCount()
	response.ControlPassedCount = calc.ControlPassedCount()
	response.FindingCount = calc.FindingCount()
	log.Printf("%d controls (%d Passed and %d Failed)", calc.total, calc.passed, calc.failed)
	log.Printf("Compliance score is: %.2f%%", response.Score)

	return response, err
}

func (x *Lambda) downloadControls(bucket string, key string) ([]string, error) {
	var controls []string

	data, err := x.downloadFile(bucket, key)

	if err != nil {
		return controls, err
	}

	err = json.Unmarshal(data, &controls)
	log.Printf("Downloaded %d controls", len(controls))

	return controls, err
}

func (x *Lambda) downloadFindings(bucket string, key string) ([]*Finding, error) {
	var findings []*Finding

	data, err := x.downloadFile(bucket, key)

	if err != nil {
		return []*Finding{}, err
	}

	err = json.Unmarshal(data, &findings)
	log.Printf("Downloaded %d findings", len(findings))

	return findings, err
}

func (x *Lambda) downloadFile(bucket string, key string) ([]byte, error) {
	log.Printf("Downloading s3://%s/%s", bucket, key)

	response, err := x.s3Client.GetObject(x.ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(response.Body)
}
