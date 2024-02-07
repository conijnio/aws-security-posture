package main

type Request struct {
	AccountId string   `json:"AccountId"`
	Bucket    string   `json:"Bucket"`
	Key       string   `json:"Key"`
	GroupBy   string   `json:"GroupBy"`
	Controls  []string `json:"Controls"`
}

type Response struct {
	AccountId   string   `json:"AccountId"`
	Bucket      string   `json:"Bucket"`
	Key         string   `json:"Key"`
	GroupBy     string   `json:"GroupBy"`
	Controls    []string `json:"Controls"`
	Workload    string   `json:"Workload"`
	Environment string   `json:"Environment"`
}
