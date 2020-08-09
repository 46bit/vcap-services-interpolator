package cf_instance_identity

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

type Server struct {
	serverCertPath string
	serverKeyPath  string
	clientCA       *x509.CertPool
}

func NewServer(
	serverCertPath string,
	serverKeyPath string,
	instanceIdentityCA []byte,
) *Server {
	clientCA := x509.NewCertPool()
	clientCA.AppendCertsFromPEM(instanceIdentityCA)
	return &Server{serverCertPath, serverKeyPath, clientCA}
}

func (s *Server) ServeWithMutualTLS(httpServer *http.Server) error {
	tlsConfig := &tls.Config{
		ClientCAs:  s.clientCA,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()
	httpServer.TLSConfig = tlsConfig
	return httpServer.ListenAndServeTLS(s.serverCertPath, s.serverKeyPath)
}
