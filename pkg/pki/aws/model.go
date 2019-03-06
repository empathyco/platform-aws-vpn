package awspki

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/empathybroker/aws-vpn/pkg/pki"
	"github.com/pkg/errors"
)

const (
	kAttrSerialNumber   = "SerialNumber"
	kAttrAuthorityKeyId = "AuthorityKeyId"
	kAttrSubjectKeyId   = "SubjectKeyId"
	kAttrSubjectName    = "SubjectName"
	kAttrCertType       = "CertType"
	kAttrIssuedAt       = "IssuedAt"
	kAttrValidUntil     = "ValidUntil"
	kAttrRevocationTime = "RevocationTime"
	kAttrData           = "Data"

	kIndexSubjectKeyId = kAttrSubjectKeyId + "Idx"
	kIndexSubjectName  = kAttrSubjectName + "Idx"
)

type dynamoCertEntry struct {
	SerialNumber   []byte    `dynamodbav:",binary"`
	AuthorityKeyId []byte    `dynamodbav:",binary"`
	SubjectKeyId   []byte    `dynamodbav:",binary"`
	SubjectName    string    `dynamodbav:",string"`
	CertType       string    `dynamodbav:",string"`
	IssuedAt       time.Time `dynamodbav:",unixtime"`
	ValidUntil     time.Time `dynamodbav:",unixtime"`
	RevocationTime time.Time `dynamodbav:",unixtime"`
	Data           []byte    `dynamodbav:",binary"`
}

func (e *dynamoCertEntry) toCertificateInfo() (*pki.CertificateInfo, error) {
	if e.Data == nil {
		return nil, nil
	}

	cert, err := x509.ParseCertificate(e.Data)
	if err != nil {
		return nil, errors.Wrap(err, "parsing certificate")
	}

	info := &pki.CertificateInfo{
		Certificate: cert,
		SerialBytes: e.SerialNumber,

		CertType:  pki.CertType(e.CertType),
		Serial:    hex.EncodeToString(e.SerialNumber),
		KeyId:     hex.EncodeToString(e.SubjectKeyId),
		Subject:   e.SubjectName,
		NotBefore: e.IssuedAt.UTC(),
		NotAfter:  e.ValidUntil.UTC(),
	}

	if e.RevocationTime.Unix() > 0 {
		rt := e.RevocationTime.UTC()
		info.Revoked = &rt
	}

	return info, err
}

func (e dynamoCertEntry) String() string {
	vals := []string{
		fmt.Sprintf("SerialNumber:%s", hex.EncodeToString(e.SerialNumber)),
		fmt.Sprintf("AuthorityKeyId:%s", hex.EncodeToString(e.AuthorityKeyId)),
		fmt.Sprintf("SubjectKeyId:%s", hex.EncodeToString(e.SubjectKeyId)),
		fmt.Sprintf("SubjectName:%s", e.SubjectName),
		fmt.Sprintf("CertType:%s", e.CertType),
		fmt.Sprintf("IssuedAt:%s", e.IssuedAt),
		fmt.Sprintf("ValidUntil:%s", e.ValidUntil),
		fmt.Sprintf("RevocationTime:%s", e.RevocationTime),
	}

	return fmt.Sprintf("{%s}", strings.Join(vals, ", "))
}

type RevokedCert struct {
	SerialNumber   []byte
	RevocationTime time.Time
}
