package main

type Request struct {
	Report             string   `json:"Report"`
	Timestamp          int64    `json:"Timestamp"`
	Bucket             string   `json:"Bucket"`
	GroupBy            string   `json:"GroupBy"`
	Controls           []string `json:"Controls"`
	NextToken          string   `json:"NextToken"`
	FindingCount       int      `json:"FindingCount"`
	Findings           []string `json:"Findings"`
	AggregatedFindings []string `json:"AggregatedFindings"`
}

type Account struct {
	AccountId string   `json:"AccountId"`
	Bucket    string   `json:"Bucket"`
	Key       string   `json:"Key"`
	GroupBy   string   `json:"GroupBy"`
	Controls  []string `json:"Controls"`
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
	Report    string    `json:"Report"`
	Timestamp int64     `json:"Timestamp"`
	Bucket    string    `json:"Bucket"`
	Accounts  []Account `json:"Accounts"`
}
