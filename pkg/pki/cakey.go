package pki

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	jose "gopkg.in/square/go-jose.v2"
)

type CAData struct {
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey

	CACert     *x509.Certificate
	PrevCACert *x509.Certificate
	CrossCert  *x509.Certificate

	StaticKey StaticKey
}

type storedCAData struct {
	PrivateKey jose.JSONWebKey `json:"key"`

	CACert     []byte `json:"ca"`
	PrevCACert []byte `json:"pca,omitempty"`
	CrossCert  []byte `json:"xca,omitempty"`
	StaticKey  []byte `json:"ovpn"`
}

func (k *CAData) UnmarshalJSON(data []byte) error {
	var stored storedCAData
	if err := json.Unmarshal(data, &stored); err != nil {
		return errors.Wrap(err, "unmarshal CA data")
	}

	parsed, err := x509.ParseCertificate(stored.CACert)
	if err != nil {
		return errors.Wrap(err, "parsing CA certificate")
	}
	k.CACert = parsed

	if len(stored.PrevCACert) > 0 {
		k.PrevCACert, err = x509.ParseCertificate(stored.PrevCACert)
		if err != nil {
			return errors.Wrap(err, "parsing old CA certificate")
		}
	}

	if len(stored.CrossCert) > 0 {
		k.CrossCert, err = x509.ParseCertificate(stored.CrossCert)
		if err != nil {
			return errors.Wrap(err, "parsing cross-signed CA certificate")
		}
	}

	var ok bool
	if k.PrivateKey, ok = stored.PrivateKey.Key.(crypto.PrivateKey); !ok {
		return errors.New("unexpected privateKey type")
	}

	if k.PublicKey, ok = stored.PrivateKey.Public().Key.(crypto.PublicKey); !ok {
		return errors.New("unexpected publicKey type")
	}

	k.StaticKey = stored.StaticKey

	return nil
}

func (k CAData) MarshalJSON() ([]byte, error) {
	s := storedCAData{
		PrivateKey: jose.JSONWebKey{Key: k.PrivateKey},
		CACert:     k.CACert.Raw,
		StaticKey:  k.StaticKey,
	}

	if k.PrevCACert != nil {
		s.PrevCACert = k.PrevCACert.Raw
	}

	if k.CrossCert != nil {
		s.CrossCert = k.CrossCert.Raw
	}

	return json.Marshal(s)
}

func NewCAKey(caName string, serialNumber string, duration time.Duration) (CAData, error) {
	privKey, err := NewPrivateKey()
	if err != nil {
		return CAData{}, errors.Wrap(err, "error generating key")
	}

	pkiName := pkix.Name{CommonName: caName, SerialNumber: serialNumber}
	caCert, err := CreateCertificate(nil, privKey, GetPublicKey(privKey), pkiName, CACert, WithDuration(duration), WithMaxPathLen(1))
	if err != nil {
		return CAData{}, errors.Wrap(err, "error signing certificate key")
	}

	return CAData{
		PrivateKey: privKey,
		PublicKey:  GetPublicKey(privKey),
		CACert:     caCert,
		PrevCACert: nil,
		CrossCert:  nil,
		StaticKey:  NewStaticKey(),
	}, nil
}

func (k CAData) Renew(caName string, serialNumber string, duration time.Duration) (CAData, error) {
	privKey, err := NewPrivateKey()
	if err != nil {
		return CAData{}, errors.Wrap(err, "error generating key")
	}

	pkiName := pkix.Name{CommonName: caName, SerialNumber: serialNumber}
	caCert, err := CreateCertificate(nil, privKey, GetPublicKey(privKey), pkiName, CACert, WithDuration(duration), WithMaxPathLen(1))
	if err != nil {
		return CAData{}, errors.Wrap(err, "error signing CA certificate")
	}

	crossCert, err := CreateCertificate(k.CACert, k.PrivateKey, GetPublicKey(privKey), pkiName, CACert, WithExpiration(k.CACert.NotAfter), WithMaxPathLen(0))
	if err != nil {
		return CAData{}, errors.Wrap(err, "error cross-signing CA certificate")
	}

	return CAData{
		PrivateKey: privKey,
		PublicKey:  GetPublicKey(privKey),
		CACert:     caCert,
		PrevCACert: k.CACert,
		CrossCert:  crossCert,
		StaticKey:  k.StaticKey,
	}, nil
}
