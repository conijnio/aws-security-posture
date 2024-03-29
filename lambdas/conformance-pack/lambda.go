package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofrs/uuid"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Lambda struct {
	ctx      context.Context
	client   *configservice.Client
	s3Client *s3.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.client = configservice.NewFromConfig(cfg)
	m.s3Client = s3.NewFromConfig(cfg)
	return m
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	x.ctx = ctx
	response := Response{
		Report:   request.Report,
		Bucket:   request.Bucket,
		Controls: x.resolveBucketKey("controls", request.Report),
		GroupBy:  "Title",
		Filter:   request.Filter,
	}

	log.Printf("Loading Conformance Pack Context: %s", request.ConformancePack)
	controls, err := x.resolveConfigRules(request.ConformancePack)
	controlsData, err := json.Marshal(controls)

	if err != nil {
		return response, err
	}

	err = x.uploadFile(request.Bucket, response.Controls, controlsData)
	return response, err
}

func (x *Lambda) resolveConfigRules(conformancePack string) ([]string, error) {
	paginator := configservice.NewGetConformancePackComplianceDetailsPaginator(x.client, &configservice.GetConformancePackComplianceDetailsInput{
		ConformancePackName: aws.String(conformancePack),
		Limit:               100,
	})

	var encountered = map[string]bool{}

	pageNum := 0
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []string{}, err
		}
		for _, value := range output.ConformancePackRuleEvaluationResults {
			parts := strings.Split(*value.EvaluationResultIdentifier.EvaluationResultQualifier.ConfigRuleName, "-")
			control := strings.Join(parts[:len(parts)-1], "-")
			encountered[control] = true
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
