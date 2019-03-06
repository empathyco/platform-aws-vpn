package gsuite

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/aws/aws-xray-sdk-go/xray"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
)

type GoogleServiceAccountSecretProvider interface {
	GetKey(context.Context) ([]byte, error)
}

type GoogleDirectory struct {
	secretProvider GoogleServiceAccountSecretProvider

	service *admin.Service
	once    sync.Once
}

type UserInfo struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`

	IsAdmin     bool `json:"is_admin"`
	IsSuspended bool `json:"is_suspended"`

	Schemas map[string]map[string]interface{} `json:"schemas"`
}

func NewGoogleDirectory(secretProvider GoogleServiceAccountSecretProvider) *GoogleDirectory {
	return &GoogleDirectory{
		secretProvider: secretProvider,
	}
}

func (d *GoogleDirectory) init(ctx context.Context) {
	d.once.Do(func() {
		jsonKey, err := d.secretProvider.GetKey(ctx)
		if err != nil {
			log.WithError(err).Fatal("Error obtaining Google Service Account key")
		}

		config, err := google.JWTConfigFromJSON(jsonKey, admin.AdminDirectoryUserReadonlyScope)
		if err != nil {
			log.WithError(err).Fatal("Error creating config from Service Account key")
		}

		if iSubject, ok := os.LookupEnv("GSUITE_IMPERSONATE_SUBJECT"); ok {
			config.Subject = iSubject
		}

		oaClient := config.Client(ctx)
		oaClient = xray.Client(oaClient)
		d.service, err = admin.New(oaClient)
		if err != nil {
			log.WithError(err).Fatal("Error creating Admin SDK client")
		}
	})
}

func (d *GoogleDirectory) GetUserInfo(ctx context.Context, userKey string) (*UserInfo, error) {
	d.init(ctx)

	res, err := d.service.Users.Get(userKey).Projection("full").Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	schemas := make(map[string]map[string]interface{})
	for s, r := range res.CustomSchemas {
		schemas[s] = make(map[string]interface{})
		var schema map[string]interface{}
		if err := json.Unmarshal(r, &schema); err == nil {
			for k, v := range schema {
				if v1, ok := v.([]interface{}); ok {
					var values []interface{}
					for _, v2 := range v1 {
						if v3, ok := v2.(map[string]interface{}); ok {
							if v4, ok := v3["value"]; ok {
								values = append(values, v4)
							}
						} else {
							values = append(values, v3)
						}
					}
					schemas[s][k] = values
				} else {
					schemas[s][k] = v
				}
			}
		}
	}

	return &UserInfo{
		Id:    res.Id,
		Email: res.PrimaryEmail,
		Name:  res.Name.FullName,

		IsAdmin:     res.IsAdmin,
		IsSuspended: res.Suspended,

		Schemas: schemas,
	}, nil
}
