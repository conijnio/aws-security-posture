package main

type Request struct {
	GeneratorId string `json:"generator-id"`
	Report      string `json:"report"`
	Bucket      string `json:"bucket"`
}

type Response struct {
	Report    string `json:"Report"`
	Bucket    string `json:"Bucket"`
	Key       string `json:"Key"`
	Timestamp int64  `json:"Timestamp"`
}
