package main

type Request struct {
	Report    string    `json:"Report"`
	Timestamp int64     `json:"Timestamp"`
	Bucket    string    `json:"Bucket"`
	Accounts  []Account `json:"Accounts"`
}

type Account struct {
	AccountId   string `json:"AccountId"`
	AccountName string `json:"AccountName"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
	GroupBy     string `json:"GroupBy"`
	Controls    string `json:"Controls"`
}

type Response struct {
	Report    string    `json:"Report"`
	Timestamp int64     `json:"Timestamp"`
	Bucket    string    `json:"Bucket"`
	Accounts  []Account `json:"Accounts"`
}
