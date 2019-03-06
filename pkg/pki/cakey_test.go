package pki

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	if err := os.Setenv("PKI_KEY_TYPE", "RSA"); err != nil {
		panic(err)
	}
}

const (
	kCAName   = "EmpathyBroker VPN"
	kDuration = 90 * 24 * time.Hour
)

func TestNewCAKey(t *testing.T) {
	if err := os.Setenv("PKI_KEY_TYPE", "RSA"); err != nil {
		t.Fatal(err)
	}

	caKey, err := NewCAKey(kCAName, uuid.New().String(), kDuration)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	rCaKey, err := caKey.Renew(kCAName, uuid.New().String(), kDuration)
	if err != nil {
		t.Fatal(err)
	}

	caData, err := json.Marshal(caKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(caData), string(caData))

	rCaData, err := json.Marshal(rCaKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(rCaData), string(rCaData))
}

func TestRenewSize(t *testing.T) {
	maxSize := 7 * 1024

	max := []int{0, 0, 0}

	for i := 0; i < 10; i++ {
		caKey, err := NewCAKey(kCAName, uuid.New().String(), kDuration)
		if err != nil {
			t.Fatal(err)
		}

		caData, err := json.Marshal(caKey)
		if err != nil {
			t.Fatal(err)
		}

		if len(caData) > max[0] {
			max[0] = len(caData)
		}
		// t.Log(0, len(caData))

		caKey, err = caKey.Renew(kCAName, uuid.New().String(), kDuration)
		if err != nil {
			t.Fatal(err)
		}

		caData, err = json.Marshal(caKey)
		if err != nil {
			t.Fatal(err)
		}

		if len(caData) > max[1] {
			max[1] = len(caData)
		}
		t.Log(1, len(caData), string(caData))

		if len(caData) > maxSize {
			t.Fatalf("%d > %d", len(caData), maxSize)
		}
	}

	t.Log(max)

}
