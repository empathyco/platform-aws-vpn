package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/coreos/go-systemd/dbus"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context/ctxhttp"
)

const (
	kJsonContentType = "application/json"
)

var (
	sess   = session.Must(session.NewSession())
	client = &http.Client{
		Transport: awsservices.NewAWSSigner(sess, "execute-api", http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	apiBase = os.Getenv("PKI_API_ENDPOINT")
)

func init() {
	if !strings.HasPrefix(apiBase, "https://") {
		log.Fatalf("Missing or incorrect PKI_API_ENDPOINT environment variable")
	}
}

func hexFromEnv(key string) string {
	return strings.ReplaceAll(os.Getenv(key), ":", "")
}

func intFromEnv(key string) int {
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.WithError(err).Errorf("Error decoding env int: %s=%s", key, os.Getenv(key))
	}
	return val
}

func apiRequest(ctx context.Context, method string, url string, params map[string]interface{}, result interface{}) (int, error) {
	var body bytes.Buffer
	if params != nil {
		if err := json.NewEncoder(&body).Encode(params); err != nil {
			return 0, errors.Wrap(err, "could not marshal request")
		}
	}

	req, err := http.NewRequest(method, apiBase+url, &body)
	req.Header.Set("Content-Type", kJsonContentType)

	res, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return 0, errors.Wrapf(err, "error querying service %s", url)
	}
	defer res.Body.Close()

	if result != nil {
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return res.StatusCode, errors.Wrap(err, "error unmarshaling response")
		}
	}

	return res.StatusCode, nil
}

func restartService(ctx context.Context, unit string) error {
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)

	conn, err := dbus.NewSystemConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := make(chan string)
	if _, err := conn.ReloadOrRestartUnit(unit, "replace", ch); err != nil {
		return err
	}

	select {
	case result := <-ch:
		switch result {
		case "done", "skipped":
			log.Debugf("Reloaded service %s", unit)
			return nil
		default:
			return fmt.Errorf("error reloading: %s", result)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
