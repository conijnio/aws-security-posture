package main

type Request struct {
	AccountId string `json:"AccountId"`
	Bucket    string `json:"Bucket"`
	Key       string `json:"Key"`
	GroupBy   string `json:"GroupBy"`
}

type Response struct {
	AccountId   string `json:"AccountId"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
	GroupBy     string `json:"GroupBy"`
	Workload    string `json:"Workload"`
	Environment string `json:"Environment"`
}
