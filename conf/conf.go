package conf

// Conf config struct
type Conf struct {
	Addr       string `json:"addr,omitempty"`
	MgrAddr    string `json:"mgrAddr,omitempty"`
	EtcdAddr   string
	EtcdPrefix string

	LogLevel string `json:"logLevel,omitempty"`

	EnableRuntimeVal bool `json:"enableRuntimeVal,omitempty"`
	EnableCookieVal  bool `json:"enableCookieVal,omitempty"`
	EnableHeadVal    bool `json:"enableHeadVal,omitempty"`

	ReqHeadStaticMapping map[string]string `json:"reqHeadStaticMapping,omitempty"`
	RspHeadStaticMapping map[string]string `json:"rspHeadStaticMapping,omitempty"`
}
