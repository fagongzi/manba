package conf

import (
	"encoding/json"
	"io/ioutil"

	"github.com/fagongzi/log"
)

// Conf config struct
type Conf struct {
	Addr    string `json:"addr"`
	MgrAddr string `json:"mgrAddr"`

	RegistryAddr string `json:"registryAddr"`
	Prefix       string `json:"prefix"`

	Filers []*FilterSpec `json:"filers"`

	// MaxServerCheckSec max check server interval seconds
	MaxServerCheckSec int `json:"maxServerCheckSec"`
	// Maximum number of connections which may be established to server
	MaxConns int `json:"maxConns"`
	// MaxConnDuration Keep-alive connections are closed after this duration.
	MaxConnDuration int `json:"maxConnDuration"`
	// MaxIdleConnDuration Idle keep-alive connections are closed after this duration.
	MaxIdleConnDuration int `json:"maxIdleConnDuration"`
	// ReadBufferSize Per-connection buffer size for responses' reading.
	ReadBufferSize int `json:"readBufferSize"`
	// WriteBufferSize Per-connection buffer size for requests' writing.
	WriteBufferSize int `json:"writeBufferSize"`
	// ReadTimeout Maximum duration for full response reading (including body).
	ReadTimeout int `json:"readTimeout"`
	// WriteTimeout Maximum duration for full request writing (including body).
	WriteTimeout int `json:"writeTimeout"`
	// MaxResponseBodySize Maximum response body size.
	MaxResponseBodySize int `json:"maxResponseBodySize"`

	// EnablePPROF enable pprof
	EnablePPROF bool `json:"enablePPROF"`
	// PPROFAddr pprof addr
	PPROFAddr string `json:"pprofAddr,omitempty"`
}

// FilterSpec filter spec
type FilterSpec struct {
	Name               string `json:"name"`
	External           bool   `json:"external,omitempty"`
	ExternalPluginFile string `json:"externalPluginFile,omitempty"`
}

// GetCfg returns the conf from external file
func GetCfg(file string) *Conf {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("bootstrap: read config file <%s> failed, errors:\n%+v",
			file,
			err)
	}

	cnf := &Conf{}
	err = json.Unmarshal(data, cnf)
	if err != nil {
		log.Fatalf("bootstrap: parse config file <%s> failed, errors:\n%+v",
			file,
			err)
	}

	return cnf
}
