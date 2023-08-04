package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
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
		AccountId: request.AccountId,
		Score:     0,
	}
	x.ctx = ctx
	log.Printf("Calculating the security score for: %s", request.AccountId)

	findings, err := x.downloadFindings(request.Bucket, request.Key)

	if err != nil {
		return response, err
	}

	failed := 0
	notAvailable := 0
	passed := 0
	warning := 0

	for _, finding := range findings {
		if finding.Compliance.Status == types.ComplianceStatusPassed {
			passed++
		}
		if finding.Compliance.Status == types.ComplianceStatusNotAvailable {
			passed++
			notAvailable++
		}
		if finding.Compliance.Status == types.ComplianceStatusFailed {
			failed++
		}
		if finding.Compliance.Status == types.ComplianceStatusWarning {
			failed++
			warning++
		}
	}

	response.Score = (float64(passed) / float64(len(findings))) * 100
	log.Printf("%d Passed findings (%d in NotAvailable)", passed, notAvailable)
	log.Printf("%d Failed findings (%d in Warning)", failed, warning)
	log.Printf("Compliance score is: %.2f%%", response.Score)

	return response, err
}

func (x *Lambda) downloadFindings(bucket string, key string) ([]*types.AwsSecurityFinding, error) {
	data, err := x.downloadFile(bucket, key)

	if err != nil {
		return []*types.AwsSecurityFinding{}, err
	}

	var findings []*types.AwsSecurityFinding
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
