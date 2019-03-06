package pki

import (
	"crypto/x509"
	"time"
)

type CertOptions func(cert *x509.Certificate)

func CACert(cert *x509.Certificate) {
	cert.IsCA = true
	cert.BasicConstraintsValid = true
	cert.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	cert.ExtKeyUsage = nil
}

func ClientCert(cert *x509.Certificate) {
	cert.IsCA = false
	cert.BasicConstraintsValid = true
	cert.KeyUsage = x509.KeyUsageDigitalSignature
	cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
}

func ServerCert(cert *x509.Certificate) {
	cert.IsCA = false
	cert.BasicConstraintsValid = true
	cert.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
}

func WithTimespan(notBefore time.Time, notAfter time.Time) CertOptions {
	return func(cert *x509.Certificate) {
		cert.NotBefore = notBefore
		cert.NotAfter = notAfter
	}
}

func WithDuration(duration time.Duration) CertOptions {
	notBefore := time.Now()
	return WithTimespan(notBefore, notBefore.Add(duration))
}

func WithExpiration(notAfter time.Time) CertOptions {
	return WithTimespan(time.Now(), notAfter)
}

func WithMaxPathLen(pathLen int) CertOptions {
	return func(cert *x509.Certificate) {
		cert.MaxPathLen = pathLen
		cert.MaxPathLenZero = pathLen == 0
	}
}

func WithEmail(email ...string) CertOptions {
	return func(cert *x509.Certificate) {
		cert.EmailAddresses = email
	}
}

func WithDNS(dnsName ...string) CertOptions {
	return func(cert *x509.Certificate) {
		cert.DNSNames = dnsName
	}
}
