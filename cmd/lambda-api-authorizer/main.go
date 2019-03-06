package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/arn"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/gsuite"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "@timestamp",
		},
	})

	verifier = NewGoogleTokenVerifier(configApiAuthorizer.Audience)
}

var (
	verifier *TokenVerifier

	awsSM      = awsservices.NewSecretsManagerClient()
	gDirectory = gsuite.NewGoogleDirectory(awsservices.NewAWSServiceAccountProvider(awsSM, "VPN/GoogleServiceAccount"))
)

type googleClaims struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	HostedDomain  string `json:"hd"`
}

func (c googleClaims) verifyHostedDomain() bool {
	for _, hd := range configApiAuthorizer.HostedDomains {
		if c.HostedDomain == hd {
			return true
		}
	}
	return false
}

func authHandler(ctx context.Context, input *events.APIGatewayCustomAuthorizerRequest) (*events.APIGatewayCustomAuthorizerResponse, error) {
	if input.Type != "TOKEN" {
		return nil, errors.New("expected TOKEN authorizer")
	}

	method, err := arn.Parse(input.MethodArn)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing method ARN")
	}
	log.Debugf("method: %s", method)

	authn := strings.SplitN(input.AuthorizationToken, " ", 2)
	if len(authn) != 2 || authn[0] != "Bearer" {
		log.Error("Invalid auth header")
		return nil, errors.New("Unauthorized")
	}

	var extraClaims *googleClaims
	jwtClaims, err := verifier.VerifyToken(ctx, authn[1], &extraClaims)
	if err != nil {
		log.WithError(err).Error("Invalid token")
		return nil, errors.New("Unauthorized")
	}

	if !extraClaims.EmailVerified {
		log.Error("Unverified email")
		return nil, errors.New("Unauthorized")
	}

	if !extraClaims.verifyHostedDomain() {
		log.Errorf("invalid hosted domain: %s", extraClaims.HostedDomain)
		return nil, errors.New("Unauthorized")
	}

	userInfo, err := gDirectory.GetUserInfo(ctx, jwtClaims.Subject)
	if err != nil {
		log.WithError(err).Error("Error fetching user info")
		return nil, errors.New("Unauthorized")
	}

	for k := range userInfo.Schemas {
		if k != "VPN" {
			delete(userInfo.Schemas, k)
		}
	}

	googleInfo, err := json.Marshal(userInfo)
	if err != nil {
		log.WithError(err).Error("Error marshaling Google info")
		return nil, errors.New("Unauthorized")
	}

	authContext := make(map[string]interface{})
	authContext["email"] = extraClaims.Email
	authContext["hd"] = extraClaims.HostedDomain
	authContext["name"] = extraClaims.Name
	authContext["google"] = string(googleInfo)

	var stmts []events.IAMPolicyStatement
	stmts = append(stmts, events.IAMPolicyStatement{
		Effect:   "Allow",
		Action:   []string{"execute-api:Invoke"},
		Resource: []string{"*"},
	})

	log.Infof("Authorized: %s", extraClaims.Email)

	return &events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: jwtClaims.Subject,
		Context:     authContext,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version:   "2012-10-17",
			Statement: stmts,
		},
	}, nil
}

func main() {
	lambda.Start(authHandler)
}
