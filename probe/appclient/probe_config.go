package appclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"github.com/certifi/gocertifi"
	"github.com/weaveworks/scope/common/xfer"
)

var certPool *x509.CertPool

func init() {
	var err error
	certPool, err = gocertifi.CACerts()
	if err != nil {
		panic(err)
	}
}

// ProbeConfig contains all the info needed for a probe to do HTTP requests
type ProbeConfig struct {
	Token        string
	ProbeVersion string
	ProbeID      string
	Insecure     bool
}

func (pc ProbeConfig) authorizeHeaders(headers http.Header) {
	headers.Set("Authorization", fmt.Sprintf("Scope-Probe token=%s", pc.Token))
	headers.Set(xfer.ScopeProbeIDHeader, pc.ProbeID)
	headers.Set(xfer.ScopeProbeVersionHeader, pc.ProbeVersion)
}

func (pc ProbeConfig) authorizedRequest(method string, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err == nil {
		pc.authorizeHeaders(req.Header)
	}
	return req, err
}

func (pc ProbeConfig) getHTTPTransport(hostname string) (*http.Transport, error) {
	var tlsConfig *tls.Config
	if pc.Insecure {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		tlsConfig = &tls.Config{
			RootCAs:    certPool,
			ServerName: hostname,
		}
	}
	return &http.Transport{
		TLSClientConfig: tlsConfig,

		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		IdleConnTimeout:       http.DefaultTransport.IdleConnTimeout,
		TLSHandshakeTimeout:   http.DefaultTransport.TLSHandshakeTimeout,
		ExpectContinueTimeout: http.DefaultTransport.ExpectContinueTimeout,
	}
}
