package main

import "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

type Request struct {
	Report             string                          `json:"Report"`
	Bucket             string                          `json:"Bucket"`
	Controls           string                          `json:"Controls"`
	GroupBy            string                          `json:"GroupBy"`
	Filter             types.AwsSecurityFindingFilters `json:"Filter"`
	Findings           []string                        `json:"Findings"`
	FindingCount       int                             `json:"FindingCount"`
	AggregatedFindings []string                        `json:"AggregatedFindings"`
	NextToken          string                          `json:"NextToken"`
	Timestamp          int64                           `json:"Timestamp"`
}

type Response struct {
	Report             string                          `json:"Report"`
	Bucket             string                          `json:"Bucket"`
	Controls           string                          `json:"Controls"`
	GroupBy            string                          `json:"GroupBy"`
	Filter             types.AwsSecurityFindingFilters `json:"Filter"`
	Findings           []string                        `json:"Findings"`
	FindingCount       int                             `json:"FindingCount"`
	AggregatedFindings []string                        `json:"AggregatedFindings"`
	NextToken          string                          `json:"NextToken"`
	Timestamp          int64                           `json:"Timestamp"`
}

type Finding struct {
	Id             string `json:"Id"`
	Status         string `json:"Status"`
	ProductArn     string `json:"ProductArn"`
	GeneratorId    string `json:"GeneratorId"`
	AwsAccountId   string `json:"AwsAccountId"`
	AwsAccountName string `json:"AwsAccountName"`
	Title          string `json:"Title"`
}
