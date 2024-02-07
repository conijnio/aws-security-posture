package main

type Request struct {
	Report          string    `json:"Report"`
	Timestamp       int64     `json:"Timestamp"`
	Bucket          string    `json:"Bucket"`
	ConformancePack string    `json:"ConformancePack"`
	Accounts        []Account `json:"Accounts"`
}

type Account struct {
	AccountId   string `json:"AccountId"`
	AccountName string `json:"AccountName"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
}

type Response struct {
	Report          string            `json:"Report"`
	Timestamp       int64             `json:"Timestamp"`
	Bucket          string            `json:"Bucket"`
	ConformancePack string            `json:"ConformancePack"`
	Accounts        []Account         `json:"Accounts"`
	AccountMapping  map[string]string `json:"AccountMapping"`
}
