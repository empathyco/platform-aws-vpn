package pki

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"time"
)

type CertificateInfo struct {
	Certificate *x509.Certificate `json:"-"`
	SerialBytes []byte            `json:"-"`

	CertType  CertType   `json:"type"`
	Serial    string     `json:"serial"`
	KeyId     string     `json:"keyId"`
	Subject   string     `json:"subject"`
	NotBefore time.Time  `json:"notBefore"`
	NotAfter  time.Time  `json:"notAfter"`
	Revoked   *time.Time `json:"revoked,omitempty"`
}

func CertInfoFromX509Cert(cert *x509.Certificate) *CertificateInfo {
	return &CertificateInfo{
		Certificate: cert,
		SerialBytes: cert.SerialNumber.Bytes(),

		CertType:  GetCertType(cert),
		Serial:    hex.EncodeToString(cert.SerialNumber.Bytes()),
		KeyId:     hex.EncodeToString(cert.SubjectKeyId),
		Subject:   cert.Subject.CommonName,
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		Revoked:   nil,
	}
}

type PKIStorage interface {
	GetCACert(ctx context.Context) *x509.Certificate
	GetPrevCACert(ctx context.Context) *x509.Certificate
	GetCrossCert(ctx context.Context) *x509.Certificate
	GetPrivateKey(ctx context.Context) crypto.PrivateKey
	GetPublicKey(ctx context.Context) crypto.PublicKey
	GetStaticKey(ctx context.Context) StaticKey

	AddCert(ctx context.Context, cert *x509.Certificate) error
	ListAllCerts(ctx context.Context) ([]*CertificateInfo, error)
	ListCertsBySubject(context.Context, string) ([]*CertificateInfo, error)
	GetCertBySerial(context.Context, []byte) (*CertificateInfo, error)
	RevokeCert(context.Context, []byte) (*CertificateInfo, error)
}
