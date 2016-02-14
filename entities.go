package main

//Service ...
type Service struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	Protocol  string `json:"protocol"`
	Scheduler string `json:"scheduler"`
}

//Services ...
type Services []Service
