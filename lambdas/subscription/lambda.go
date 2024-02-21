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
	"path/filepath"
	"sort"
	"time"
)

type Lambda struct {
	ctx      context.Context
	client   *securityhub.Client
	s3Client *s3.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.client = securityhub.NewFromConfig(cfg)
	m.s3Client = s3.NewFromConfig(cfg)
	return m
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	x.ctx = ctx
	response := Response{
		Report:   request.Report,
		Bucket:   request.Bucket,
		Controls: x.resolveBucketKey("controls", request.Report),
		GroupBy:  "GeneratorId",
		Filter:   request.Filter,
	}

	log.Printf("Loading control based on SubscriptionArn: %s", request.SubscriptionArn)

	controls, err := x.resolveConfigRules(request.SubscriptionArn)
	controlsData, err := json.Marshal(controls)

	if err != nil {
		return response, err
	}

	err = x.uploadFile(request.Bucket, response.Controls, controlsData)

	return response, err
}

func (x *Lambda) resolveConfigRules(subscriptionArn string) ([]string, error) {

	paginator := securityhub.NewDescribeStandardsControlsPaginator(x.client, &securityhub.DescribeStandardsControlsInput{
		StandardsSubscriptionArn: aws.String(subscriptionArn),
		MaxResults:               aws.Int32(100),
	})

	var encountered = map[string]bool{}

	pageNum := 0
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []string{}, err
		}
		for _, value := range output.Controls {
			if value.ControlStatus == types.ControlStatusEnabled {
				encountered[*value.StandardsControlArn] = true
			}
		}
		pageNum++
	}

	var controls []string
	for control := range encountered {
		controls = append(controls, control)
	}

	sort.Strings(controls)

	return controls, nil
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
