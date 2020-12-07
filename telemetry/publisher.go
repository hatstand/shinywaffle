//go:generate protoc  --go_out=plugins=grpc:. telemetry.proto
package telemetry

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lestrrat/go-jwx/jwk"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
)

var projectID = flag.String("project-id", "hodoor-211bb", "Project ID for Cloud IoT")

type WrapperMessage struct {
	BinaryData []byte `json:"binary_data"`
}

type Publisher struct {
	key *rsa.PrivateKey
}

func (p *Publisher) Publish(name string, temp float64, on bool) error {
	data, err := marshal(&IOTMessage{
		IotMessage: &IOTMessage_Telemetry{
			Telemetry: &TelemetryMessage{
				Name:        name,
				Temperature: temp,
				On:          on,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to serialise message: %v", err)
	}
	return p.send(data)
}

func (p *Publisher) Hello() error {
	data, err := marshal(&IOTMessage{
		IotMessage: &IOTMessage_Hello{
			Hello: &HelloMessage{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to serialise message: %v", err)
	}
	return p.send(data)
}

func NewPublisher() *Publisher {
	keys, err := jwk.ParseString(os.Getenv("IOT_KEY"))
	if err != nil {
		log.Fatalf("Failed to parse JWK: %v", err)
	}
	if len(keys.Keys) != 1 {
		log.Fatalf("Wrong number of keys in JWK: %d", len(keys.Keys))
	}
	key, err := keys.Keys[0].Materialize()
	if err != nil {
		log.Fatalf("Failed to materialize key: %v", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("Failed to convert RSA key")
	}

	return &Publisher{
		key: rsaKey,
	}
}

func marshal(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proto: %v", err)
	}

	encoded, err := json.Marshal(&WrapperMessage{
		BinaryData: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wrap message: %v", err)
	}
	return encoded, nil
}

func (p *Publisher) send(data []byte) error {
	now := time.Now()
	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		&jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Minute).Unix(),
			Audience:  *projectID,
		})
	sig, err := token.SignedString(p.key)
	if err != nil {
		log.Fatalf("Failed to sign JWT: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(
		"https://cloudiotdevice.googleapis.com/v1/projects/%s/locations/%s/registries/%s/devices/%s:publishEvent",
		*projectID, "europe-west1", "hodoor", "hodoor"),
		bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Failed to build message: %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sig))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send startup message: %v", err)
	}
	return nil
}
