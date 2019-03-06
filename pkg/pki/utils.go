package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type CertSerial *big.Int
type CertType string

const (
	CertTypeUnknown CertType = "UNKNOWN"
	CertTypeServer           = "Server"
	CertTypeClient           = "Client"
	CertTypeCA               = "CA"
)

var kSerialMaxValue = new(big.Int).Lsh(big.NewInt(1), 128)

func DecodeSerial(serial string) ([]byte, error) {
	decoded, err := hex.DecodeString(serial)
	if err != nil {
		return nil, err
	}

	res := big.NewInt(0).SetBytes(decoded)
	if res.Cmp(kSerialMaxValue) > 0 {
		return nil, errors.New("serial out of range")
	}

	return res.Bytes(), nil
}

func randomSerial() CertSerial {
	serial, err := rand.Int(rand.Reader, kSerialMaxValue)
	if err != nil {
		log.WithError(err).Fatal("failed to generate serial number")
	}
	return serial
}

func getSKI(key crypto.PublicKey) ([]byte, error) {
	keyASN, err := x509.MarshalPKIXPublicKey(key)
	var spki struct {
		Algo      pkix.AlgorithmIdentifier
		BitString asn1.BitString
	}

	if rest, err := asn1.Unmarshal(keyASN, &spki); err != nil {
		return nil, errors.WithStack(err)
	} else if len(rest) > 0 {
		return nil, errors.Errorf("unexpected %d remaining bytes", len(rest))
	}

	h := sha1.Sum(spki.BitString.Bytes)
	return h[:], err
}

func EncodePEMCert(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

func EncodeDERPrivateKey(key crypto.PrivateKey) ([]byte, error) {
	switch key := key.(type) {
	case *rsa.PrivateKey:
		return x509.MarshalPKCS1PrivateKey(key), nil
	case *ecdsa.PrivateKey:
		return x509.MarshalECPrivateKey(key)
	default:
		return nil, errors.New("unsupported key type")
	}
}

func EncodePEMPrivateKey(key crypto.PrivateKey) ([]byte, error) {
	derBytes, err := EncodeDERPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "encoding private key")
	}

	var block pem.Block
	switch key.(type) {
	case *rsa.PrivateKey:
		block = pem.Block{Type: "RSA PRIVATE KEY", Bytes: derBytes}
	case *ecdsa.PrivateKey:
		block = pem.Block{Type: "EC PRIVATE KEY", Bytes: derBytes}
	default:
		return nil, errors.New("unsupported key type")
	}

	return pem.EncodeToMemory(&block), nil
}

func NewPrivateKey() (crypto.PrivateKey, error) {
	switch os.Getenv("PKI_KEY_TYPE") {
	case "RSA":
		return rsa.GenerateKey(rand.Reader, 2048)
	case "EC":
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	default:
		return nil, errors.New("invalid PKI_KEY_TYPE")
	}
}

func GetPublicKey(key crypto.PrivateKey) crypto.PublicKey {
	switch key := key.(type) {
	case *rsa.PrivateKey:
		return key.Public()
	case *ecdsa.PrivateKey:
		return key.Public()
	default:
		panic("unsupported key type")
	}
}

func GetCertType(cert *x509.Certificate) CertType {
	if cert.IsCA {
		return CertTypeCA
	}

	for _, eku := range cert.ExtKeyUsage {
		if eku == x509.ExtKeyUsageServerAuth {
			return CertTypeServer
		}
		if eku == x509.ExtKeyUsageClientAuth {
			return CertTypeClient
		}
	}

	return CertTypeUnknown
}
