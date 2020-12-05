package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/golang/protobuf/proto"
	"github.com/hatstand/shinywaffle/telemetry"
)

type pushRequest struct {
	Message struct {
		ID          string `json:"message_id"`
		Data        []byte
		Attributes  map[string]string
		PublishTime time.Time `json:"publish_time"`
	}
	Subscription string
}

func (r *pushRequest) Save() (map[string]bigquery.Value, string, error) {
	ret := make(map[string]bigquery.Value)

	var iot telemetry.IOTMessage
	if err := proto.Unmarshal(r.Message.Data, &iot); err != nil {
		return ret, "", fmt.Errorf("failed to deserialise embedded IoT message: %v", err)
	}

	switch x := iot.IotMessage.(type) {
	case *telemetry.IOTMessage_Telemetry:
		t := x.Telemetry
		ret["temperature"] = t.Temperature
		ret["name"] = t.Name
		ret["on"] = t.On
	default:
		return ret, "", fmt.Errorf("don't know how to save this: %+v", iot)
	}

	ret["timestamp"] = r.Message.PublishTime
	return ret, r.Message.ID, nil
}

func main() {
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, "hodoor-211bb")
	if err != nil {
		log.Fatalf("Coult not connect to BigQuery: %v", err)
	}

	ins := client.Dataset("telemetry").Table("device_telemetry").Inserter()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, World!")
	})

	http.HandleFunc("/telemetry/push", func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		var msg pushRequest
		if err := json.Unmarshal(data, &msg); err != nil {
			http.Error(w, fmt.Sprintf("Could not decode push request: %v", err), http.StatusBadRequest)
			return
		}

		var iot telemetry.IOTMessage
		if err := proto.Unmarshal(msg.Message.Data, &iot); err != nil {
			http.Error(w, fmt.Sprintf("Could not decode IoT message: %v", err), http.StatusBadRequest)
			return
		}

		log.Printf("IoT: %s", iot.String())

		switch x := iot.IotMessage.(type) {
		case *telemetry.IOTMessage_Telemetry:
			items := []*pushRequest{
				&msg,
			}
			if err := ins.Put(ctx, items); err != nil {
				log.Printf("failed to store message: %v", err)
				multiErr, ok := err.(bigquery.PutMultiError)
				if ok {
					for _, err := range multiErr {
						log.Printf("Failed to store row: %+v", err)
					}
				}
				http.Error(w, fmt.Sprintf("Failed to store message: %v", err), http.StatusInternalServerError)
				return
			}
		default:
			log.Printf("Skipping unrecognised message type: %v", x)
		}
	})

	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
