package main

type Request struct {
	Report             string   `json:"Report"`
	Timestamp          int64    `json:"Timestamp"`
	Bucket             string   `json:"Bucket"`
	Controls           string   `json:"Controls"`
	GroupBy            string   `json:"GroupBy"`
	Findings           []string `json:"Findings"`
	AggregatedFindings []string `json:"AggregatedFindings"`
}

type Account struct {
	AccountId   string `json:"AccountId"`
	AccountName string `json:"AccountName"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
	Controls    string `json:"Controls"`
	GroupBy     string `json:"GroupBy"`
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
	Report    string    `json:"Report"`
	Timestamp int64     `json:"Timestamp"`
	Bucket    string    `json:"Bucket"`
	Accounts  []Account `json:"Accounts"`
}
