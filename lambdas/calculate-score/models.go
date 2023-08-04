package main

type Request struct {
	AccountId string `json:"AccountId"`
	Bucket    string `json:"Bucket"`
	Key       string `json:"Key"`
}

type Response struct {
	AccountId string  `json:"AccountId"`
	Score     float64 `json:"Score"`
}
