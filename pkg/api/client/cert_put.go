package clientapi

import (
	"bytes"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/empathybroker/aws-vpn/pkg/api"

	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/ovpn"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	kMaxClientCerts = 2
)

type newCertRequest struct {
	PublicKey jose.JSONWebKey `json:"publicKey"`
}

func apiNewCert(w http.ResponseWriter, r *http.Request) {
	var request newCertRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err, "Invalid input")
		return
	}

	if !request.PublicKey.Valid() {
		api.ErrorResponse(w, http.StatusBadRequest, nil, "Invalid public key")
		return
	}

	_, userInfo, err := api.GetAPIGWPrincipal(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error obtaining principal")
		return
	}

	currentCerts, err := apiPKI.ListCerts(r.Context(), userInfo.Email)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error listing certs")
		return
	}

	nonRevoked := 0
	for _, cCert := range currentCerts {
		if cCert.Revoked == nil {
			nonRevoked += 1

			if nonRevoked >= kMaxClientCerts {
				if _, err := apiPKI.RevokeCert(r.Context(), cCert.SerialBytes); err != nil {
					log.WithError(err).Error("Error revoking certificate")
					// Don't fail
				}
			}
		}
	}

	name := pkix.Name{
		CommonName: userInfo.Email,
	}

	/*name.ExtraNames = append(name.ExtraNames, pkix.AttributeTypeAndValue{
		Type:  asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 3}, // mail
		Value: userInfo.Email},
	)*/
	name.ExtraNames = append(name.ExtraNames, pkix.AttributeTypeAndValue{
		Type:  asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 1}, // uid
		Value: userInfo.Id},
	)

	cert, err := apiPKI.CreateCertificate(r.Context(), request.PublicKey.Key, name,
		pki.WithDuration(30*24*time.Hour), pki.ClientCert, pki.WithEmail(userInfo.Email))
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error creating certificate")
		return
	}

	event := api.J{
		"event": "cert_signed",
		"cert":  cert,
	}

	if err := awsservices.PublishEvent(apiSNS, r.Context(), event); err != nil {
		log.WithError(err).Error("Error publishing event")
	}

	configData := ovpn.ConfigData{
		Certificate: cert.Certificate,

		CACert:     apiPKI.GetCACert(r.Context()),
		PrevCACert: apiPKI.GetPrevCACert(r.Context()),
		CrossCert:  apiPKI.GetCrossCert(r.Context()),

		StaticKey: apiPKI.GetStaticKey(r.Context()),
	}

	var buf bytes.Buffer
	if err := ovpn.GetClientConfig(&buf, configData); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err, "Error writing OpenVPN profile")
		return
	}

	fname := "OpenVPN"
	if caName := os.Getenv("PKI_CLIENT_CERT_NAME"); caName != "" {
		fname = caName
	}

	w.Header().Set("X-VPN-Filename", fmt.Sprintf("%s.ovpn", fname))
	w.Header().Set("Content-Type", "application/x-openvpn-profile")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.WithError(err).Error("Error writing binary response")
	}
}
