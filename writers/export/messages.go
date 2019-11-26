// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/cisco/senml"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/logger"
	s "github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

const pointName = "messages"

var _ writers.MessageRepository = (*exportRepo)(nil)

type exportRepo struct {
	client mqtt.Client
	conf   Config
	log    logger.Logger
}

type Config struct {
	NatsURL        string
	LogLevel       string
	Port           string
	MqttHost       string
	MqttUsername   string
	MqttPassword   string
	MqttMTLS       bool
	MqttSkipTLSVer bool
	MqttRetain     bool
	MqttChannel    string
	MqttCA         []byte
	MqttCert       tls.Certificate
	MqttChan       string
	MqttQoS        int
	Channels       map[string]bool
}

type fields map[string]interface{}
type tags map[string]string

// New returns new InfluxDB writer.
func New(client mqtt.Client, conf Config, log logger.Logger) writers.MessageRepository {
	return &exportRepo{
		client: client,
		conf:   conf,
		log:    log,
	}
}

func (repo *exportRepo) Save(messages ...interface{}) error {
	topic := fmt.Sprintf("channels/%s/messages/%s", repo.conf.MqttChannel)
	for _, msg := range messages {
		m, ok := msg.(senml.SenML)
		if !ok {
			return fmt.Errorf("Wrong type")
		}
		payload, err := senml.Encode(m, senml.JSON, senml.OutputOptions{})
		if err != nil {
			repo.log.Error(fmt.Sprintf("Failed to publish message on topic %s : %s", repo.conf.MqttChannel, err.Error()))
		}
		if token := repo.client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			repo.log.Error(fmt.Sprintf("Failed to publish message on topic %s : %s", repo.conf.MqttChannel, token.Error()))
		}
	}
	return nil
}

func (repo *exportRepo) tagsOf(msg *s.Message) tags {
	return tags{
		"channel":   msg.Channel,
		"subtopic":  msg.Subtopic,
		"publisher": msg.Publisher,
		"name":      msg.Name,
	}
}

func (repo *exportRepo) fieldsOf(msg *s.Message) fields {
	updateTime := strconv.FormatFloat(msg.UpdateTime, 'f', -1, 64)
	ret := fields{
		"protocol":   msg.Protocol,
		"unit":       msg.Unit,
		"link":       msg.Link,
		"updateTime": updateTime,
	}

	switch {
	case msg.Value != nil:
		ret["value"] = *msg.Value
	case msg.StringValue != nil:
		ret["stringValue"] = *msg.StringValue
	case msg.DataValue != nil:
		ret["dataValue"] = *msg.DataValue
	case msg.BoolValue != nil:
		ret["boolValue"] = *msg.BoolValue
	}

	if msg.Sum != nil {
		ret["sum"] = *msg.Sum
	}

	return ret
}
