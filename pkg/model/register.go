package model

// Register register
type Register interface {
	Registry(proxyInfo *ProxyInfo) error

	GetProxies() ([]*ProxyInfo, error)

	ChangeLogLevel(proxyAddr string, level string) error

	AddAnalysisPoint(proxyAddr, serverAddr string, secs int) error

	GetAnalysisPoint(proxyAddr, serverAddr string, secs int) (*GetAnalysisPointRsp, error)
}
