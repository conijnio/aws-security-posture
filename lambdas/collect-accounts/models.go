package main

type Request struct {
	Report    string `json:"Report"`
	Timestamp int64  `json:"Timestamp"`
	Bucket    string `json:"Bucket"`
	Key       string `json:"Key"`
}

type Account struct {
	AccountId string `json:"AccountId"`
	Bucket    string `json:"Bucket"`
	Key       string `json:"Key"`
}

type Response struct {
	Report    string    `json:"Report"`
	Timestamp int64     `json:"Timestamp"`
	Bucket    string    `json:"Bucket"`
	Accounts  []Account `json:"Accounts"`
}
