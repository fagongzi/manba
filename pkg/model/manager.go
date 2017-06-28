package model

// SetLogReq SetLogReq
type SetLogReq struct {
	Level string
}

// SetLogRsp SetLogRsp
type SetLogRsp struct {
	Code int
}

// SetReqHeadStaticMappingReq SetReqHeadStaticMappingReq
type SetReqHeadStaticMappingReq struct {
	Name  string
	Value string
}

// SetReqHeadStaticMappingRsp SetReqHeadStaticMappingRsp
type SetReqHeadStaticMappingRsp struct {
	Code int
}

// AddAnalysisPointReq AddAnalysisPointReq
type AddAnalysisPointReq struct {
	Addr string
	Secs int
}

// AddAnalysisPointRsp AddAnalysisPointRsp
type AddAnalysisPointRsp struct {
	Code int
}

// GetAnalysisPointReq GetAnalysisPointReq
type GetAnalysisPointReq struct {
	Addr string
	Secs int
}

// GetAnalysisPointRsp GetAnalysisPointRsp
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
