//go:generate protoc  --go_out=plugins=grpc:. telemetry.proto
package telemetry

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

var iotKey = flag.String("iot-key", "rsa-private.pem", "Path to RSA private key file for Cloud IoT")
var projectID = flag.String("project-id", "hodoor-211bb", "Project ID for Cloud IoT")

type Publisher struct {
	mq *client.Client
}

func (p *Publisher) Publish(name string, temp float64, on bool) error {
	msg := &TelemetryMessage{
		Name:        name,
		Temperature: temp,
		On:          on,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialise telemetry message: %v", err)
	}
	err = p.mq.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		TopicName: []byte("projects/hodoor-211bb/topics/heating"),
		Message:   data,
	})
	if err != nil {
		return fmt.Errorf("failed to publish telemetry message: %v", err)
	}
	return nil
}

func (p *Publisher) Close() {
	p.mq.Terminate()
}

func NewPublisher() *Publisher {
	mq := client.New(&client.Options{
		ErrorHandler: func(err error) {
			log.Printf("MQTT error: %v", err)
		},
	})
	defer mq.Terminate()

	keyData, err := ioutil.ReadFile(*iotKey)
	if err != nil {
		log.Fatalf("Failed to load IoT private key: %v", err)
	}
	pemBlock, _ := pem.Decode(keyData)
	if pemBlock == nil || pemBlock.Type != "PRIVATE KEY" {
		log.Fatalf("Failed to parse PEM: %v", err)
	}
	key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse RSA key: %v", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("Failed to extract RSA key")
	}

	now := time.Now()
	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		&jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Minute).Unix(),
			Audience:  *projectID,
		})
	sig, err := token.SignedString(rsaKey)
	if err != nil {
		log.Fatalf("Failed to sign JWT: %v", err)
	}

	err = mq.Connect(&client.ConnectOptions{
		Network:   "tcp",
		Address:   "mqtt.googleapis.com:8883",
		ClientID:  []byte("projects/hodoor-211bb/locations/europe-west1/registries/hodoor/devices/hodoor"),
		TLSConfig: &tls.Config{},
		UserName:  []byte("ignored"),
		Password:  []byte(sig),
	})
	if err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	err = mq.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		TopicName: []byte("projects/hodoor-211bb/topics/heating"),
		Message:   []byte("hello"),
	})
	if err != nil {
		log.Fatalf("Failed to publish hello message: %v", err)
	}
	log.Printf("Published startup message")

	return &Publisher{
		mq: mq,
	}
}
