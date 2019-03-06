package awspki

import (
	"context"
	"crypto/x509"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	A "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	E "github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	log "github.com/sirupsen/logrus"
)

func (s *awsStorage) GetCertBySerial(ctx context.Context, serial []byte) (*pki.CertificateInfo, error) {
	res, err := s.ddb.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(configAWSPKI.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			kAttrSerialNumber: {B: serial},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	var certEntry dynamoCertEntry
	if err := A.UnmarshalMap(res.Item, &certEntry); err != nil {
		return nil, err
	}

	return certEntry.toCertificateInfo()
}

func (s *awsStorage) ListAllCerts(ctx context.Context) ([]*pki.CertificateInfo, error) {
	filter := E.GreaterThan(E.Key(kAttrValidUntil), E.Value(time.Now().UTC().Unix()))
	filter = filter.And(E.Equal(E.Key(kAttrCertType), E.Value(pki.CertTypeClient)))

	exp, err := E.NewBuilder().
		WithFilter(filter).
		Build()
	if err != nil {
		return nil, err
	}

	query := &dynamodb.ScanInput{
		TableName: aws.String(configAWSPKI.TableName),

		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		FilterExpression:          exp.Filter(),
	}

	certs := make([]*pki.CertificateInfo, 0)
	if err := s.ddb.ScanPagesWithContext(ctx, query, func(output *dynamodb.ScanOutput, b bool) bool {
		for _, item := range output.Items {
			var certEntry dynamoCertEntry
			if err := A.UnmarshalMap(item, &certEntry); err != nil {
				log.WithError(err).Error("Error unmarshaling cert from Dynamo")
				continue
			}

			info, err := certEntry.toCertificateInfo()
			if err != nil {
				log.WithError(err).Error("Error parsing certificate from Dynamo")
				continue
			}

			certs = append(certs, info)
		}

		return b
	}); err != nil {
		return nil, err
	}

	return certs, nil
}

func (s *awsStorage) ListCertsBySubject(ctx context.Context, subjectName string) ([]*pki.CertificateInfo, error) {
	exp, err := E.NewBuilder().
		WithKeyCondition(E.KeyEqual(E.Key(kAttrSubjectName), E.Value(subjectName))).
		WithFilter(E.GreaterThan(E.Key(kAttrValidUntil), E.Value(time.Now().UTC().Unix()))).
		Build()
	if err != nil {
		return nil, err
	}

	query := dynamodb.QueryInput{
		TableName: aws.String(configAWSPKI.TableName),
		IndexName: aws.String(kIndexSubjectName),

		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
		FilterExpression:          exp.Filter(),

		ScanIndexForward: aws.Bool(false),
	}

	certs := make([]*pki.CertificateInfo, 0)
	if err := s.ddb.QueryPagesWithContext(ctx, &query, func(output *dynamodb.QueryOutput, b bool) bool {
		for _, item := range output.Items {
			var certEntry dynamoCertEntry
			if err := A.UnmarshalMap(item, &certEntry); err != nil {
				log.WithError(err).Errorf("error unmarshaling cert")
				continue
			}

			info, err := certEntry.toCertificateInfo()
			if err != nil {
				log.WithError(err).Errorf("error parsing certificate")
				continue
			}

			certs = append(certs, info)
		}

		return b
	}); err != nil {
		return nil, err
	}

	return certs, nil
}

func (s *awsStorage) AddCert(ctx context.Context, cert *x509.Certificate) error {
	if cert.Raw == nil {
		return errors.New("missing cert raw data")
	}

	if cert.AuthorityKeyId == nil {
		return errors.New("missing Authority Key ID")
	}

	if cert.SubjectKeyId == nil {
		return errors.New("missing Subject Key ID")
	}

	cType := pki.GetCertType(cert)
	if cType == pki.CertTypeUnknown {
		return errors.New("unknown certificate type")
	}

	entry := dynamoCertEntry{
		SerialNumber:   cert.SerialNumber.Bytes(),
		AuthorityKeyId: cert.AuthorityKeyId,
		SubjectKeyId:   cert.SubjectKeyId,
		SubjectName:    cert.Subject.CommonName,
		CertType:       string(cType),
		IssuedAt:       cert.NotBefore.UTC(),
		ValidUntil:     cert.NotAfter.UTC(),
		RevocationTime: time.Unix(0, 0).UTC(),
		Data:           cert.Raw,
	}

	item, err := A.MarshalMap(entry)
	if err != nil {
		return err
	}

	_, err = s.ddb.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(configAWSPKI.TableName),
		Item:      item,
	})

	return err
}

func (s *awsStorage) RevokeCert(ctx context.Context, serial []byte) (*pki.CertificateInfo, error) {
	expr, err := E.NewBuilder().
		WithCondition(E.Equal(E.Name(kAttrRevocationTime), E.Value(0))).
		WithUpdate(E.Set(E.Name(kAttrRevocationTime), E.Value(time.Now().UTC().Unix()))).
		Build()
	if err != nil {
		return nil, err
	}

	res, err := s.ddb.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(configAWSPKI.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			kAttrSerialNumber: {B: serial},
		},

		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),

		ReturnValues: aws.String(dynamodb.ReturnValueAllNew),
	})
	if err != nil {
		return nil, err
	}

	var certEntry dynamoCertEntry
	if err := A.UnmarshalMap(res.Attributes, &certEntry); err != nil {
		return nil, err
	}

	return certEntry.toCertificateInfo()
}
