package main

type CalculatedScore struct {
	AccountId string  `json:"AccountId"`
	Score     float64 `json:"Score"`
}

type Request struct {
	Report    string             `json:"Report"`
	Timestamp int64              `json:"Timestamp"`
	Bucket    string             `json:"Bucket"`
	Accounts  []*CalculatedScore `json:"Accounts"`
}

type Response struct {
}
