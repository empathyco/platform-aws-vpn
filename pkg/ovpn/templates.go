package ovpn

import (
	"crypto/x509"
	"io"
	"os"
	"text/template"

	"github.com/empathybroker/aws-vpn/pkg/pki"
)

var (
	funcs = template.FuncMap{
		"pemCert": pki.EncodePEMCert,
		"env":     os.Getenv,
	}
	tplClientConfig = template.Must(template.New("client_config").Funcs(funcs).Parse(kClientConfigTemplate))
	tplServerConfig = template.Must(template.New("server_config").Funcs(funcs).Parse(kServerConfigTemplate))
)

type ConfigData struct {
	Certificate *x509.Certificate

	CACert     *x509.Certificate
	PrevCACert *x509.Certificate
	CrossCert  *x509.Certificate

	StaticKey pki.StaticKey
}

func GetClientConfig(w io.Writer, data ConfigData) error {
	return tplClientConfig.Execute(w, data)
}

func GetServerConfig(w io.Writer, data ConfigData) error {
	return tplServerConfig.Execute(w, data)
}
