package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/empathybroker/aws-vpn/pkg/pki"
	log "github.com/sirupsen/logrus"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	kConfigLocationEnv = "OPENVPN_CONFIG_FILE"
	kServiceUnitEnv    = "OPENVPN_SERVICE_UNIT"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	})

	if verb, ok := os.LookupEnv("verb"); ok {
		if verb, err := strconv.Atoi(verb); err != nil && verb > 3 {
			log.SetLevel(log.DebugLevel)
		}
	}
}

func getServerConfig(ctx context.Context) {
	log.Debugf("Fetching server certificate")

	configFileName, ok := os.LookupEnv(kConfigLocationEnv)
	if !ok {
		log.Fatalf("Missing %s environment variable", kConfigLocationEnv)
	}

	privKey, err := pki.NewPrivateKey()
	if err != nil {
		log.WithError(err).Fatal("error generating key")
	}

	params := map[string]interface{}{
		"publicKey": jose.JSONWebKey{Key: pki.GetPublicKey(privKey)},
	}

	var result struct {
		Message string `json:"message"`
		Config  []byte `json:"config"`
	}

	status, err := apiRequest(ctx, http.MethodPost, "/server/config", params, &result)
	if err != nil {
		log.WithError(err).Fatalf("Error making service call")
	}

	if status != http.StatusOK {
		log.WithField("status", strconv.Itoa(status)).Fatalf("HTTP error: %s", result.Message)
	}

	encodedKey, err := pki.EncodePEMPrivateKey(privKey)
	if err != nil {
		log.WithError(err).Fatal("Error encoding private key")
	}

	configData := bytes.ReplaceAll(result.Config, []byte("%PRIVATEKEY%"), encodedKey)
	if err := ioutil.WriteFile(configFileName, configData, 0600); err != nil {
		log.WithError(err).Fatal("Could not save new configuration file")
	}

	if serviceUnit, ok := os.LookupEnv(kServiceUnitEnv); ok {
		if err := restartService(ctx, serviceUnit); err != nil {
			log.WithError(err).Fatal("Error restarting OpenVPN service")
		}
	}

	log.Exit(0)
}

func tlsVerify(ctx context.Context) {
	if len(os.Args) != 3 {
		log.Fatalf("Invalid arguments")
	}

	pathLength, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.WithError(err).Fatal("Fist parameter must be a number")
	}

	if pathLength > 2 {
		log.Fatalf("Certificate path length is too high (got %d, expected <= 2)", pathLength)
	}

	if pathLength != 0 {
		// We want to wait for the leaf cert
		log.Exit(0)
	}

	log.Debugf("Validating certificate for %s", os.Args[2])

	params := map[string]interface{}{
		"subject":      os.Args[2],
		"untrusted_ip": os.Getenv("untrusted_ip"),

		"serial": hexFromEnv("tls_serial_hex_0"),
		"digest": hexFromEnv("tls_digest_sha256_0"),
	}

	var result struct {
		Message string `json:"message"`
	}

	status, err := apiRequest(ctx, http.MethodPost, "/server/verify", params, &result)
	if err != nil {
		log.WithError(err).Fatalf("Error making service call")
	}

	if status != http.StatusOK {
		log.WithField("status", strconv.Itoa(status)).Fatalf("HTTP error: %s", result.Message)
	}

	log.Debugf("Certificate validation successful!")
	log.Exit(0)
}

func clientConnect(ctx context.Context) {
	if len(os.Args) != 2 {
		log.Fatalf("Invalid arguments")
	}

	params := map[string]interface{}{
		"common_name": os.Getenv("common_name"),
		"trusted_ip":  os.Getenv("trusted_ip"),

		"client_hwaddr":   os.Getenv("IV_HWADDR"),
		"client_platform": os.Getenv("IV_PLAT"),
		"client_version":  os.Getenv("IV_VER"),
		"client_gui":      os.Getenv("IV_GUI"),
		"client_ssl":      os.Getenv("IV_SSL"),
	}

	var result struct {
		Message string   `json:"message"`
		Push    []string `json:"push"`
	}

	status, err := apiRequest(ctx, http.MethodPost, "/server/connect", params, &result)
	if err != nil {
		log.WithError(err).Fatalf("Error making service call")
	}

	if status != http.StatusOK {
		log.WithField("status", strconv.Itoa(status)).Fatalf("HTTP error: %s", result.Message)
	}

	var configData bytes.Buffer
	for _, p := range result.Push {
		if _, err := fmt.Fprintf(&configData, "push \"%s\"\n", p); err != nil {
			log.WithError(err).Fatal("Error writing config data")
		}
	}

	if err := ioutil.WriteFile(os.Args[1], configData.Bytes(), 0600); err != nil {
		log.WithError(err).Fatal("Could not save client configuration file")
	}

	log.Exit(0)
}

func clientDisconnect(ctx context.Context) {
	if len(os.Args) != 1 {
		log.Fatalf("Invalid arguments")
	}

	params := map[string]interface{}{
		"common_name": os.Getenv("common_name"),
		"trusted_ip":  os.Getenv("trusted_ip"),

		"duration":       intFromEnv("time_duration"),
		"bytes_sent":     intFromEnv("bytes_sent"),
		"bytes_received": intFromEnv("bytes_received"),

		"client_hwaddr":   os.Getenv("IV_HWADDR"),
		"client_platform": os.Getenv("IV_PLAT"),
		"client_version":  os.Getenv("IV_VER"),
		"client_gui":      os.Getenv("IV_GUI"),
		"client_ssl":      os.Getenv("IV_SSL"),
	}

	var result struct {
		Message string   `json:"message"`
		Push    []string `json:"push"`
	}

	status, err := apiRequest(ctx, http.MethodPost, "/server/disconnect", params, &result)
	if err != nil {
		log.WithError(err).Fatalf("Error making service call")
	}

	if status != http.StatusOK {
		log.WithField("status", strconv.Itoa(status)).Fatalf("HTTP error: %s", result.Message)
	}

	log.Exit(0)
}

func main() {
	ctx := context.Background()

	if scriptType, ok := os.LookupEnv("script_type"); ok {
		switch scriptType {
		case "tls-verify":
			tlsVerify(ctx)
		case "client-connect":
			clientConnect(ctx)
		case "client-disconnect":
			clientDisconnect(ctx)
		default:
			log.Fatalf("Unknown script_type: %s", scriptType)
		}
		log.Exit(1)
	}

	if len(os.Args) < 2 {
		log.Fatalf("Invalid arguments")
	}

	switch os.Args[1] {
	case "getconfig":
		getServerConfig(ctx)
	default:
		log.Fatalf("Unknown command %s", os.Args[1])
	}

	log.Exit(1)
}
