package cloudstack

type queryAsyncJobResultResponse struct {
	QueryAsyncJobResultResponse struct {
		JobStatus     int         `json:"jobstatus"`
		JobResult     interface{} `json:"jobresult"`
		JobResultType string      `json:"jobresulttype"`
		JobResultCode int         `json:"jobresultcode"`
	} `json:"queryasyncjobresultresponse"`
}
