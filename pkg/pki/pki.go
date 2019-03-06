package pki

import (
	"context"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"

	"github.com/pkg/errors"
)

type PKI struct {
	storage PKIStorage
}

func NewPKI(s PKIStorage) *PKI {
	return &PKI{
		storage: s,
	}
}

func (pki *PKI) GetCACert(ctx context.Context) *x509.Certificate {
	return pki.storage.GetCACert(ctx)
}

func (pki *PKI) GetPrevCACert(ctx context.Context) *x509.Certificate {
	return pki.storage.GetPrevCACert(ctx)
}

func (pki *PKI) GetCrossCert(ctx context.Context) *x509.Certificate {
	return pki.storage.GetCrossCert(ctx)
}

func (pki *PKI) GetStaticKey(ctx context.Context) StaticKey {
	return pki.storage.GetStaticKey(ctx)
}

func (pki *PKI) CreateCertificate(ctx context.Context, pubKey crypto.PublicKey, subject pkix.Name, certOpts ...CertOptions) (*CertificateInfo, error) {
	cert, err := CreateCertificate(pki.storage.GetCACert(ctx), pki.storage.GetPrivateKey(ctx), pubKey, subject, certOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "creating certificate")
	}

	if err := pki.storage.AddCert(ctx, cert); err != nil {
		return nil, errors.Wrap(err, "storing certificate")
	}

	return CertInfoFromX509Cert(cert), nil
}

func (pki *PKI) GetCertBySerial(ctx context.Context, serial []byte) (*CertificateInfo, error) {
	return pki.storage.GetCertBySerial(ctx, serial)
}

func (pki *PKI) ListCerts(ctx context.Context, subject string) ([]*CertificateInfo, error) {
	if subject == "" {
		return pki.storage.ListAllCerts(ctx)
	}
	return pki.storage.ListCertsBySubject(ctx, subject)
}

func (pki *PKI) RevokeCert(ctx context.Context, serial []byte) (*CertificateInfo, error) {
	return pki.storage.RevokeCert(ctx, serial)
}
