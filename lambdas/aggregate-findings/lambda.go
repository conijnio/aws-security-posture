package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/gofrs/uuid"
	"io"
	"log"
	"path/filepath"
	"time"
)

type Lambda struct {
	ctx               context.Context
	s3Client          *s3.Client
	securityHubClient *securityhub.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.securityHubClient = securityhub.NewFromConfig(cfg)
	m.s3Client = s3.NewFromConfig(cfg)
	return m
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	x.ctx = ctx
	log.Printf("Running a report for: %s", request.Report)
	log.Printf("Use the '%s' bucket", request.Bucket)

	aggregatedFindings, err := x.aggregateFindings(request.Bucket, request.Findings)

	if err != nil {
		return Response{}, err
	}

	findingsData, err := json.Marshal(aggregatedFindings)

	if err != nil {
		return Response{}, err
	}

	objectKey := x.resolveBucketKey("aggregated", request.Report)
	err = x.uploadFile(request.Bucket, objectKey, findingsData)

	return Response{
		Report:             request.Report,
		Bucket:             request.Bucket,
		GroupBy:            request.GroupBy,
		Controls:           request.Controls,
		Filter:             request.Filter,
		FindingCount:       0,
		Findings:           []string{},
		AggregatedFindings: append(request.AggregatedFindings, objectKey),
		NextToken:          request.NextToken,
		Timestamp:          time.Now().Unix(),
	}, err
}

func (x *Lambda) aggregateFindings(bucket string, findings []string) ([]Finding, error) {
	var aggregatedFindings []Finding

	for _, finding := range findings {
		data, err := x.downloadFile(bucket, finding)

		if err != nil {
			return aggregatedFindings, err
		}

		var findings []Finding
		err = json.Unmarshal(data, &findings)

		if err != nil {
			return aggregatedFindings, err
		}

		aggregatedFindings = append(aggregatedFindings, findings...)
	}

	return aggregatedFindings, nil
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

func (x *Lambda) uploadFile(bucket string, key string, data []byte) error {
	log.Printf("Upload file to s3://%s/%s", bucket, key)

	_, err := x.s3Client.PutObject(x.ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	viewData := string(data)
	log.Printf(viewData)

	return err
}

func (x *Lambda) resolveBucketKey(prefix string, report string) string {
	t := time.Now()
	id, _ := uuid.NewV6()

	return filepath.Join(
		report,
		prefix,
		fmt.Sprintf("%d", t.Year()),
		fmt.Sprintf("%02d", int(t.Month())),
		fmt.Sprintf("%02d", t.Day()),
		fmt.Sprintf("%s.json", id.String()),
	)
}
