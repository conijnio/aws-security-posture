package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
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

	findings, err := x.downloadFindings(request.Bucket, request.Key)

	if err != nil {
		return response, err
	}

	for accountId, accountFindings := range x.splitPerAccountId(findings) {
		data, _ := json.Marshal(accountFindings)
		accountObjectKey, err := x.uploadFile(accountId, data)

		if err != nil {
			return response, err
		}

		response.Accounts = append(response.Accounts, Account{
			AccountId: accountId,
			Bucket:    request.Bucket,
			Key:       accountObjectKey,
		})
	}

	return response, err
}

func (x *Lambda) splitPerAccountId(findings []*types.AwsSecurityFinding) map[string][]*types.AwsSecurityFinding {
	var findingsPerAccount = make(map[string][]*types.AwsSecurityFinding)

	for _, finding := range findings {
		findingsPerAccount[*finding.AwsAccountId] = append(findingsPerAccount[*finding.AwsAccountId], finding)
	}

	return findingsPerAccount
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
		fmt.Sprintf("%d", t.Day()),
		fmt.Sprintf("%d.json", t.Unix()),
	)
}
