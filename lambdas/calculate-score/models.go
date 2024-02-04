package main

type Request struct {
	AccountId   string `json:"AccountId"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
	GroupBy     string `json:"GroupBy"`
	Workload    string `json:"Workload"`
	Environment string `json:"Environment"`
}

type Finding struct {
	Id           string `json:"Id"`
	Status       string `json:"Status"`
	ProductArn   string `json:"ProductArn"`
	GeneratorId  string `json:"GeneratorId"`
	AwsAccountId string `json:"AwsAccountId"`
	Title        string `json:"Title"`
}

type Response struct {
	AccountId    string  `json:"AccountId"`
	Workload     string  `json:"Workload"`
	Environment  string  `json:"Environment"`
	Score        float64 `json:"Score"`
	ControlCount int     `json:"ControlCount"`
	FindingCount int     `json:"FindingCount"`
}
