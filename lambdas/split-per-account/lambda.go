package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"path/filepath"
	"time"
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
	x.ctx = ctx
	x.ctx = context.WithValue(ctx, "request", request)

	response := Response{
		Report:    request.Report,
		Bucket:    request.Bucket,
		Accounts:  []Account{},
		Timestamp: request.Timestamp,
	}

	aggregatedFindings, err := x.downloadFindings(request.Bucket, request.AggregatedFindings)

	if err != nil {
		return response, err
	}

	findings, err := x.downloadFindings(request.Bucket, request.Findings)

	if err != nil {
		return response, err
	}

	mergedFindings := append(aggregatedFindings, findings...)

	for accountId, accountFindings := range x.splitPerAccountId(mergedFindings) {
		data, _ := json.Marshal(accountFindings)
		accountObjectKey, err := x.uploadFile(accountId, data)

		if err != nil {
			return response, err
		}

		response.Accounts = append(response.Accounts, Account{
			AccountId: accountId,
			Bucket:    request.Bucket,
			Key:       accountObjectKey,
			GroupBy:   request.GroupBy,
			Controls:  request.Controls,
		})
	}

	return response, err
}

func (x *Lambda) splitPerAccountId(findings []*Finding) map[string][]*Finding {
	var findingsPerAccount = make(map[string][]*Finding)

	for _, finding := range findings {
		AwsAccountId := finding.AwsAccountId
		findingsPerAccount[AwsAccountId] = append(findingsPerAccount[AwsAccountId], finding)
	}

	return findingsPerAccount
}

func (x *Lambda) downloadFindings(bucket string, keys []string) ([]*Finding, error) {
	var findings []*Finding

	for _, key := range keys {
		data, err := x.downloadFile(bucket, key)
		if err != nil {
			return []*Finding{}, err
		}
		var records []*Finding
		err = json.Unmarshal(data, &records)
		if err != nil {
			return []*Finding{}, err
		}
		findings = append(findings, records...)

	}

	log.Printf("Downloaded %d findings", len(findings))
	return findings, nil
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

func (x *Lambda) uploadFile(accountId string, data []byte) (string, error) {
	request := x.ctx.Value("request").(Request)
	key := x.resolveBucketKey(accountId)
	log.Printf("Upload to s3://%s/%s", request.Bucket, key)

	_, err := x.s3Client.PutObject(x.ctx, &s3.PutObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return key, err
}

func (x *Lambda) resolveBucketKey(accountId string) string {
	request := x.ctx.Value("request").(Request)
	t := time.Unix(request.Timestamp, 0)

	return filepath.Join(
		request.Report,
		accountId,
		fmt.Sprintf("%d", t.Year()),
		fmt.Sprintf("%02d", int(t.Month())),
		fmt.Sprintf("%02d", t.Day()),
		fmt.Sprintf("%d.json", t.Unix()),
	)
}
