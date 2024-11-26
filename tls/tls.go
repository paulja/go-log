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

var (
	//go:embed root-client.pem
	UserRootCert []byte
	//go:embed root-client-key.pem
	UserRootKey []byte
	//go:embed nobody-client.pem
	UserNobodyCert []byte
	//go:embed nobody-client-key.pem
	UserNobodyKey []byte
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

func ClientConfig(user string) (*tls.Config, error) {
	cert, err := loadUserCerts(user)
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

func loadUserCerts(name string) (cert tls.Certificate, err error) {
	switch name {
	case "root":
		cert, err = tls.X509KeyPair(UserRootCert, UserRootKey)
	case "nobody":
		cert, err = tls.X509KeyPair(UserNobodyCert, UserNobodyKey)
	case "":
		cert, err = tls.X509KeyPair(ClientCert, ClientKey)
	}
	return cert, err
}
