package main

//Service ...
type Service struct {
	Host      string `json:"host"`
	Port      uint16 `json:"port"`
	Protocol  string `json:"protocol"`
	Scheduler string `json:"scheduler"`
}

//Services ...
type Services []Service
