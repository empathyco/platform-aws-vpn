package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/empathybroker/aws-vpn/pkg/gsuite"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	accessor core.RequestAccessor
)

type J map[string]interface{}

func GetAPIGWPrincipal(r *http.Request) (string, *gsuite.UserInfo, error) {
	ctx, err := accessor.GetAPIGatewayContext(r)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}

	pid, ok := ctx.Authorizer["principalId"].(string)
	if !ok || pid == "" {
		return "", nil, errors.New("principalID not found")
	}

	googleInfo, ok := ctx.Authorizer["google"].(string)
	if !ok || googleInfo == "" {
		return "", nil, errors.New("Google data not found")
	}

	var userInfo *gsuite.UserInfo
	if err := json.Unmarshal([]byte(googleInfo), &userInfo); err != nil {
		return "", nil, errors.Wrap(err, "unmarshaling Google data")
	}

	return pid, userInfo, nil
}

func JsonResponse(w http.ResponseWriter, statusCode int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.WithError(err).Errorf("Error writing JSON response")
	}
}

func ErrorResponse(w http.ResponseWriter, statusCode int, err error, msg string) {
	e := J{"message": msg}

	if err != nil && os.Getenv("DEBUG") == "true" {
		e["error"] = err.Error()
		e["cause"] = errors.Cause(err)
	}

	log.WithError(err).Errorf("Error response with code %d: %s", statusCode, msg)
	JsonResponse(w, statusCode, e)
}
