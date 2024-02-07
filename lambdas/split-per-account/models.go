package main

type Request struct {
	Report             string   `json:"Report"`
	Timestamp          int64    `json:"Timestamp"`
	Bucket             string   `json:"Bucket"`
	ConformancePack    string   `json:"ConformancePack"`
	NextToken          string   `json:"NextToken"`
	FindingCount       int      `json:"FindingCount"`
	Findings           []string `json:"Findings"`
	AggregatedFindings []string `json:"AggregatedFindings"`
}

type Account struct {
	AccountId string `json:"AccountId"`
	Bucket    string `json:"Bucket"`
	Key       string `json:"Key"`
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
	Report          string    `json:"Report"`
	Timestamp       int64     `json:"Timestamp"`
	Bucket          string    `json:"Bucket"`
	ConformancePack string    `json:"ConformancePack"`
	Accounts        []Account `json:"Accounts"`
}
