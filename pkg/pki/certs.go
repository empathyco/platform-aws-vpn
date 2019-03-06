package pki

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
)

func CreateCertificate(parent *x509.Certificate, privKey crypto.PrivateKey, pubKey crypto.PublicKey, subject pkix.Name, certOpts ...CertOptions) (*x509.Certificate, error) {
	ski, err := getSKI(pubKey)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		Subject:      subject,
		SerialNumber: randomSerial(),
		PublicKey:    pubKey,
		SubjectKeyId: ski,
	}

	for _, opt := range certOpts {
		opt(template)
	}

	if parent == nil {
		parent = template
	}

	certData, err := x509.CreateCertificate(rand.Reader, template, parent, pubKey, privKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certData)
}
