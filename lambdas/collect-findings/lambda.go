package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
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
	log.Printf("Use the %s generator", request.GeneratorId)
	log.Printf("Use the %s bucket", request.Bucket)

	allFindings, err := x.downloadFindings(request.GeneratorId)
	if err != nil {
		return Response{}, err
	}

	objectKey := x.resolveBucketKey("raw", request.Report)
	err = x.uploadFile(request.Bucket, objectKey, allFindings)

	return Response{
		Report:    request.Report,
		Bucket:    request.Bucket,
		Key:       objectKey,
		Timestamp: time.Now().Unix(),
	}, err
}

func (x *Lambda) getFindingsForStandard(standard string) *securityhub.GetFindingsInput {
	log.Printf("Fetching all findings that have a GeneratorId that starts with: %s", standard)
	return &securityhub.GetFindingsInput{
		Filters: &types.AwsSecurityFindingFilters{
			GeneratorId: []types.StringFilter{
				{
					Comparison: "PREFIX",
					Value:      aws.String(standard),
				},
			},
			RecordState: []types.StringFilter{
				{
					Comparison: "NOT_EQUALS",
					Value:      aws.String("ARCHIVED"),
				},
			},
			WorkflowStatus: []types.StringFilter{
				{
					Comparison: "NOT_EQUALS",
					Value:      aws.String("SUPPRESSED"),
				},
			},
		},
		MaxResults: int32(100),
	}
}

func (x *Lambda) downloadFindings(standard string) ([]byte, error) {
	var allFindings []types.AwsSecurityFinding

	paginator := securityhub.NewGetFindingsPaginator(x.securityHubClient, x.getFindingsForStandard(standard))
	pageNum := 0
	for paginator.HasMorePages() && pageNum < 3 {
		page, err := paginator.NextPage(x.ctx)
		if err != nil {
			return nil, err
		}

		allFindings = append(allFindings, page.Findings...)
		pageNum++
	}

	log.Printf("Found %d findings", len(allFindings))
	return json.Marshal(allFindings)
}

func (x *Lambda) uploadFile(bucket string, key string, data []byte) error {
	log.Printf("Upload file to s3://%s/%s", bucket, key)

	_, err := x.s3Client.PutObject(x.ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return err
}

func (x *Lambda) resolveBucketKey(prefix string, report string) string {
	t := time.Now()
	return filepath.Join(
		report,
		prefix,
		fmt.Sprintf("%d", int(t.Year())),
		fmt.Sprintf("%02d", int(t.Month())),
		fmt.Sprintf("%d", int(t.Day())),
		fmt.Sprintf("%d.json", t.Unix()),
	)
}
