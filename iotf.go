package main

import (
	"encoding/json"
	"errors"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type iotfClient struct {
	credentials map[string]string
	mqttClient  *MQTT.MqttClient
	deviceType  string
	devicesSeen map[string]struct{}
}

func (self *iotfClient) Publish(device, message string) {
	if _, deviceFound := self.devicesSeen[device]; deviceFound == false {
		err := self.registerDevice(device)
		if err != nil {
			logger.Error("Could not register device: " + err.Error())
		}
	}
	mqttMessage := MQTT.NewMessage([]byte(message))
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", self.deviceType, device)
	logger.Info("Publishing '%v' to %v", message, topic)
	self.mqttClient.PublishMessage(topic, mqttMessage)
}

func (self *iotfClient) registerDevice(deviceId string) error {
	registerUrl := fmt.Sprintf("https://internetofthings.ibmcloud.com/api/v0001/organizations/%v/devices", self.credentials["org"])
	body := strings.NewReader(fmt.Sprintf(`{"type": "%v", "id": "%v"}`, self.deviceType, deviceId))
	request, err := http.NewRequest("POST", registerUrl, body)
	if err != nil {
		return err
	}

	request.SetBasicAuth(self.credentials["user"], self.credentials["password"])
	request.Header.Add("Content-Type", "application/json")
	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case http.StatusForbidden:
		return errors.New("Did not autenticate successfully to IoTF")
	case http.StatusConflict:
		logger.Warning("Tried to register device that already exists: " + parseErrorFromIotf(response))
		self.devicesSeen[deviceId] = struct{}{}
		return nil
	case http.StatusCreated:
		self.devicesSeen[deviceId] = struct{}{}
		return nil
	default:
		return errors.New("Could not register device: " + parseErrorFromIotf(response))
	}
}

func parseErrorFromIotf(response *http.Response) string {
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "Did not read full response: " + err.Error()
	}

	parsedResponse := struct {
		Message string
	}{}

	logger.Debug("Response with code %v from IoTF: %v", response.Status, string(responseBody))

	err = json.Unmarshal(responseBody, &parsedResponse)
	if err != nil {
		return "JSON parsing of response failed: " + err.Error()
	}
	return parsedResponse.Message
}

func connectToIotf(iotfCreds map[string]string, deviceType string) *iotfClient {
	clientOpts := MQTT.NewClientOptions()
	clientOpts.AddBroker(iotfCreds["uri"])
	clientOpts.SetClientId(fmt.Sprintf("a:%v:$v", iotfCreds["org"], generateClientIdSuffix()))
	clientOpts.SetUsername(iotfCreds["user"])
	clientOpts.SetPassword(iotfCreds["password"])

	clientOpts.SetOnConnectionLost(func(client *MQTT.MqttClient, err error) {
		logger.Error("IoTF connection lost handler called: " + err.Error())
	})

	MQTT.WARN = log.New(os.Stdout, "", 0)
	MQTT.ERROR = log.New(os.Stdout, "", 0)
	//MQTT.DEBUG = log.New(os.Stdout, "", 0)

	mqttClient := MQTT.NewClient(clientOpts)
	_, err := mqttClient.Start()
	if err != nil {
		logger.Error(err.Error())
	}

	devicesSeen := make(map[string]struct{})
	iotfClient := &iotfClient{mqttClient: mqttClient, credentials: iotfCreds, deviceType: deviceType, devicesSeen: devicesSeen}
	return iotfClient
}

func generateClientIdSuffix() string {
	rand.Seed(time.Now().UTC().UnixNano())
	suffix := rand.Intn(1000)
	return string(suffix)
}

func extractIotfCreds(services string) map[string]string {
	servicesJson := make(map[string]interface{})
	err := json.Unmarshal([]byte(services), &servicesJson)
	if err != nil {
		logger.Error(fmt.Sprintf("%v (probably missing configuration)", err))
	}

	iotfBindings := servicesJson["iotf-service"].([]interface{})
	if err != nil {
		logger.Error(err.Error())
	}
	iotf := iotfBindings[0].(map[string]interface{})

	iotfCreds := iotf["credentials"].(map[string]interface{})
	conf := make(map[string]string)
	conf["user"] = iotfCreds["apiKey"].(string)
	conf["password"] = iotfCreds["apiToken"].(string)
	conf["uri"] = fmt.Sprintf("tls://%v:%v", iotfCreds["mqtt_host"], iotfCreds["mqtt_s_port"])
	conf["org"] = iotfCreds["org"].(string)
	return conf
}
