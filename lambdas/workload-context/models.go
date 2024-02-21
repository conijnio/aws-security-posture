package main

type Request struct {
	AccountId   string `json:"AccountId"`
	AccountName string `json:"AccountName"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
	GroupBy     string `json:"GroupBy"`
	Controls    string `json:"Controls"`
}

type Response struct {
	AccountId   string `json:"AccountId"`
	AccountName string `json:"AccountName"`
	Workload    string `json:"Workload"`
	Environment string `json:"Environment"`
	Bucket      string `json:"Bucket"`
	Key         string `json:"Key"`
	GroupBy     string `json:"GroupBy"`
	Controls    string `json:"Controls"`
}
