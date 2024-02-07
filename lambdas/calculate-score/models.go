package main

type Request struct {
	AccountId   string   `json:"AccountId"`
	AccountName string   `json:"AccountName"`
	Workload    string   `json:"Workload"`
	Environment string   `json:"Environment"`
	Bucket      string   `json:"Bucket"`
	Key         string   `json:"Key"`
	GroupBy     string   `json:"GroupBy"`
	Controls    []string `json:"Controls"`
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

type Response struct {
	AccountId          string  `json:"AccountId"`
	AccountName        string  `json:"AccountName"`
	Workload           string  `json:"Workload"`
	Environment        string  `json:"Environment"`
	Score              float64 `json:"Score"`
	ControlCount       int     `json:"ControlCount"`
	FindingCount       int     `json:"FindingCount"`
	ControlFailedCount int     `json:"ControlFailedCount"`
	ControlPassedCount int     `json:"ControlPassedCount"`
}
