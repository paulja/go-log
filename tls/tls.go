package tls

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
)

var (
	//go:embed ca.pem
	CACert []byte
)

var (
	//go:embed server.pem
	ServerCert []byte
	//go:embed server-key.pem
	ServerKey []byte
)

var (
	//go:embed client.pem
	ClientCert []byte
	//go:embed client-key.pem
	ClientKey []byte
)

func ServerConfig() (*tls.Config, error) {
	cert, err := tls.X509KeyPair(ServerCert, ServerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert pair: %v", err)
	}
	ca := x509.NewCertPool()
	if ok := ca.AppendCertsFromPEM(CACert); !ok {
		return nil, fmt.Errorf("failed to load server CA cert")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    ca,
		ServerName:   "localhost",
	}
	return tlsConfig, nil
}

func ClientConfig() (*tls.Config, error) {
	cert, err := tls.X509KeyPair(ClientCert, ClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert pair: %v", err)
	}
	ca := x509.NewCertPool()
	if ok := ca.AppendCertsFromPEM(CACert); !ok {
		return nil, fmt.Errorf("failed to load client CA cert")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
		ServerName:   "localhost",
	}
	return tlsConfig, nil
}
