package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"log"
	"strings"
)

type Lambda struct {
	ctx    context.Context
	client *configservice.Client
}

func New(cfg aws.Config) *Lambda {
	m := new(Lambda)
	m.client = configservice.NewFromConfig(cfg)
	return m
}

func (x *Lambda) Handler(ctx context.Context, request Request) (Response, error) {
	x.ctx = ctx
	response := Response{
		Report:    request.Report,
		Bucket:    request.Bucket,
		Timestamp: request.Timestamp,
		Accounts:  []Account{},
	}

	accountMapping := x.copyAccountMapping(request.AccountMapping)
	log.Printf("Loading Conformance Pack Context: %s", request.ConformancePack)
	groupBy := "Title"
	controls, err := x.resolveConfigRules(request.ConformancePack)

	if err != nil {
		return response, err
	}

	for _, account := range request.Accounts {
		response.Accounts = append(response.Accounts, Account{
			AccountId:   account.AccountId,
			AccountName: x.resolveAccountName(request, account),
			Bucket:      account.Bucket,
			Key:         account.Key,
			GroupBy:     groupBy,
			Controls:    controls,
		})

		delete(accountMapping, account.AccountId)
	}

	for accountId, accountName := range accountMapping {
		response.Accounts = append(response.Accounts, Account{
			AccountId:   accountId,
			AccountName: accountName,
			GroupBy:     groupBy,
			Controls:    controls,
		})
	}

	return response, nil
}

func (x *Lambda) copyAccountMapping(accountMapping map[string]string) map[string]string {
	var accountMappingCopy = map[string]string{}

	for accountId, accountName := range accountMapping {
		accountMappingCopy[accountId] = accountName
	}
	return accountMappingCopy
}

func (x *Lambda) resolveAccountName(request Request, account Account) string {
	if value, ok := request.AccountMapping[account.AccountId]; ok {
		return value
	}
	return ""
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

	return controls, nil
}
