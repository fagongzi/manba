package model

type SetLogReq struct {
	Level string
}

type SetLogRsp struct {
	Code int
}

type SetReqHeadStaticMappingReq struct {
	Name  string
	Value string
}

type SetReqHeadStaticMappingRsp struct {
	Code int
}

type AddAnalysisPointReq struct {
	Addr string
	Secs int
}

type AddAnalysisPointRsp struct {
	Code int
}

type GetAnalysisPointReq struct {
	Addr string
	Secs int
}

type GetAnalysisPointRsp struct {
	Code                   int `json:"schema,omitempty"`
	RequestCount           int `json:"requestCount"`
	RejectCount            int `json:"rejectCount"`
	RequestSuccessedCount  int `json:"requestSuccessedCount"`
	RequestFailureCount    int `json:"requestFailureCount"`
	ContinuousFailureCount int `json:"continuousFailureCount"`
	QPS                    int `json:"qps"`
	Max                    int `json:"max"`
	Min                    int `json:"min"`
	Avg                    int `json:"avg"`
}
