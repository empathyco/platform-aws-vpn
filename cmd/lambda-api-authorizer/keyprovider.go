package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	kGoogleKeysEndpoint = "https://www.googleapis.com/oauth2/v3/certs"
)

type KeyProvider interface {
	Keys(ctx context.Context) *jose.JSONWebKeySet
}

type RemoteKeyProvider struct {
	endpoint string
	client   http.Client
	keySet   *jose.JSONWebKeySet
	expires  time.Time
}

func NewRemoteKeyProvider(endpoint string) KeyProvider {
	return &RemoteKeyProvider{
		endpoint: endpoint,
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func NewGoogleKeyProvider() KeyProvider {
	return NewRemoteKeyProvider(kGoogleKeysEndpoint)
}

func (p *RemoteKeyProvider) updateKeySet(ctx context.Context) error {
	log.Debugf("Updating JWKs from %s", p.endpoint)

	req, err := http.NewRequest(http.MethodGet, p.endpoint, http.NoBody)
	if err != nil {
		return errors.WithStack(err)
	}

	res, err := p.client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("jwk endpoint returned status %d", res.StatusCode)
	}

	var keySet *jose.JSONWebKeySet
	if err = json.NewDecoder(res.Body).Decode(&keySet); err != nil {
		return errors.Wrap(err, "error parsing jwk endpoint response")
	}
	p.keySet = keySet

	if exp := res.Header.Get("Expires"); exp != "" {
		if exptime, err := time.Parse(time.RFC1123, exp); err == nil {
			p.expires = exptime.UTC()
		} else {
			log.WithError(err).Warnf("Error parsing JWKs expiration")
		}
	}

	log.Debugf("Updated JWKs (%d keys, next update on %s)", len(p.keySet.Keys), p.expires)
	return nil
}

func (p *RemoteKeyProvider) Keys(ctx context.Context) *jose.JSONWebKeySet {
	if time.Now().UTC().After(p.expires) {
		if err := p.updateKeySet(ctx); err != nil {
			log.WithError(err).Error("Error updating JWKs")
		}
	}

	return p.keySet
}
