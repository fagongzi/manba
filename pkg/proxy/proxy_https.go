package proxy

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
	"github.com/valyala/fasthttp"
)

func (p *Proxy) enableHTTPS() bool {
	return p.cfg.DefaultTLSCert != "" && p.cfg.DefaultTLSKey != "" && p.cfg.AddrHTTPS != ""
}

func (p *Proxy) appendCertsEmbed(server *fasthttp.Server, certData []byte, keyData []byte) {
	for _, api := range p.dispatcher.apis {
		if metapb.Up == api.meta.GetStatus() && api.meta.GetUseTLS() {
			server.AppendCertEmbed(api.meta.TlsEmbedCert.CertData, api.meta.TlsEmbedCert.KeyData)
		}
	}
	server.AppendCertEmbed(certData, keyData)
}

func (p *Proxy) configTLSConfig(server *http.Server, certData []byte, keyData []byte) {
	certs := make([]tls.Certificate, 0)
	for _, api := range p.dispatcher.apis {
		if metapb.Up == api.meta.GetStatus() && api.meta.GetUseTLS() {
			cert, err := tls.X509KeyPair(api.meta.TlsEmbedCert.CertData, api.meta.TlsEmbedCert.KeyData)
			if err != nil {
				log.Errorf("api %s has invalid TLS certs", api.meta.Name)
				continue
			}
			certs = append(certs, cert)
		}
	}
	cert, _ := tls.X509KeyPair(certData, keyData)
	certs = append(certs, cert)
	server.TLSConfig.Certificates = certs
}

func (p *Proxy) mustParseDefaultTLSCert() ([]byte, []byte) {
	certData, err := ioutil.ReadFile(p.cfg.DefaultTLSCert)
	if err != nil {
		log.Fatalf("parse https cert failed with %+v", err)
	}
	keyData, err := ioutil.ReadFile(p.cfg.DefaultTLSKey)
	if err != nil {
		log.Fatalf("parse https cert failed with %+v", err)
	}
	_, err = tls.X509KeyPair(certData, keyData)
	if err != nil {
		log.Fatalf("parse https cert failed with %+v", err)
	}
	return certData, keyData
}
