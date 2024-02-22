package main

import "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

type Request struct {
	Report      string                          `json:"Report"`
	Bucket      string                          `json:"Bucket"`
	CustomRules []string                        `json:"CustomRules"`
	Filter      types.AwsSecurityFindingFilters `json:"Filter"`
}

type Response struct {
	Report   string                          `json:"Report"`
	Bucket   string                          `json:"Bucket"`
	Controls string                          `json:"Controls"`
	GroupBy  string                          `json:"GroupBy"`
	Filter   types.AwsSecurityFindingFilters `json:"Filter"`
}
