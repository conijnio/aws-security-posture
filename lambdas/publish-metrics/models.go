package main

type CalculatedScore struct {
	AccountId    string  `json:"AccountId"`
	Workload     string  `json:"Workload"`
	Environment  string  `json:"Environment"`
	Score        float64 `json:"Score"`
	ControlCount int     `json:"ControlCount"`
	FindingCount int     `json:"FindingCount"`
}

type Request struct {
	Report    string             `json:"Report"`
	Timestamp int64              `json:"Timestamp"`
	Bucket    string             `json:"Bucket"`
	Accounts  []*CalculatedScore `json:"Accounts"`
}

type Response struct {
}
