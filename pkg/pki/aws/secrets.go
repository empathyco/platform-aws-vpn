package awspki

import (
	"context"
	"crypto"
	"crypto/x509"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	log "github.com/sirupsen/logrus"
)

func (s *awsStorage) maybeUpdate(ctx context.Context) {
	if time.Now().After(s.exp) {
		log.Debug("Updating CA secrets")
		res, err := s.sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(configAWSPKI.SecretName),
		})
		if err != nil {
			log.WithError(err).Error("Fetching CA Key")
			return
		}

		if err := s.data.UnmarshalJSON(res.SecretBinary); err != nil {
			log.WithError(err).Error("Unmarshalling CA Key")
			return
		}

		s.exp = time.Now().Add(1 * time.Minute)
	}
}

func (s *awsStorage) GetCACert(ctx context.Context) *x509.Certificate {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.maybeUpdate(ctx)

	return s.data.CACert
}

func (s *awsStorage) GetPrevCACert(ctx context.Context) *x509.Certificate {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.maybeUpdate(ctx)

	return s.data.PrevCACert
}

func (s *awsStorage) GetCrossCert(ctx context.Context) *x509.Certificate {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.maybeUpdate(ctx)

	return s.data.CrossCert
}

func (s *awsStorage) GetPrivateKey(ctx context.Context) crypto.PrivateKey {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.maybeUpdate(ctx)

	return s.data.PrivateKey
}

func (s *awsStorage) GetPublicKey(ctx context.Context) crypto.PublicKey {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.maybeUpdate(ctx)

	return s.data.PublicKey
}

func (s *awsStorage) GetStaticKey(ctx context.Context) pki.StaticKey {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.maybeUpdate(ctx)

	return s.data.StaticKey
}
