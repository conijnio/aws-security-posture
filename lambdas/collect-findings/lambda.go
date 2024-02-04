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
	"github.com/gofrs/uuid"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

func resolveMaxResults() int {
	num, err := strconv.Atoi(os.Getenv("MAX_RESULTS"))

	if err != nil {
		return 100
	}

	return num
}

func (x *Lambda) resolveGroupBy(groupBy string) string {
	supportedGrouping := map[string]bool{
		"GeneratorId": true,
		"Title":       true,
	}

	if groupBy == "" {
		log.Println("No GroupBy value suppllied, we will fallback on the 'GeneratorId'!")
		return "GeneratorId"
	}

	if !supportedGrouping[groupBy] {
		log.Printf("The GroupBy value '%s' is not supported, we will fallback on the 'GeneratorId'!", groupBy)
		return "GeneratorId"
	}

	return groupBy
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	x.ctx = ctx
	log.Printf("Running a report for: %s", request.Report)
	log.Printf("Use the '%s' bucket", request.Bucket)

	var err error
	var downloadedFindings *DownloadedFinding

	downloadedFindings, err = x.downloadFindings(&request.Filter, request.NextToken)

	if err != nil {
		return Response{}, err
	}

	findings, err := json.Marshal(downloadedFindings.Findings)

	objectKey := x.resolveBucketKey("raw", request.Report)
	err = x.uploadFile(request.Bucket, objectKey, findings)
	findingsReferenceList := append(request.Findings, objectKey)

	return Response{
		Report:  request.Report,
		Bucket:  request.Bucket,
		Filter:  request.Filter,
		GroupBy: x.resolveGroupBy(request.GroupBy),
		// Add optional fields for the next iterations
		Findings:           findingsReferenceList,
		FindingCount:       len(findingsReferenceList),
		AggregatedFindings: request.AggregatedFindings,
		Timestamp:          time.Now().Unix(),
		NextToken:          downloadedFindings.NextToken,
	}, err
}

func (x *Lambda) downloadFindings(filter *types.AwsSecurityFindingFilters, token string) (*DownloadedFinding, error) {
	var awsToken *string

	if token != "" {
		awsToken = aws.String(token)
	}

	results, err := x.securityHubClient.GetFindings(x.ctx, &securityhub.GetFindingsInput{
		Filters:    filter,
		NextToken:  awsToken,
		MaxResults: aws.Int32(int32(resolveMaxResults())),
	})

	if err != nil {
		return nil, err
	}

	return x.resolveFindings(results)
}

func (x *Lambda) resolveFindings(results *securityhub.GetFindingsOutput) (*DownloadedFinding, error) {
	var nextToken string
	var allFindings []*Finding

	for _, finding := range results.Findings {
		allFindings = append(allFindings, &Finding{
			Id:           *finding.Id,
			Status:       string(finding.Compliance.Status),
			ProductArn:   *finding.ProductArn,
			GeneratorId:  *finding.GeneratorId,
			AwsAccountId: *finding.AwsAccountId,
			Title:        *finding.Title,
		})
	}

	if results.NextToken != nil {
		nextToken = *results.NextToken
	}

	return &DownloadedFinding{
		Findings:  allFindings,
		NextToken: nextToken,
	}, nil
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
