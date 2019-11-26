// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package writer

import (
	"fmt"

	"github.com/cisco/senml"
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/transformers"
	"github.com/mainflux/mainflux/writers"
	nats "github.com/nats-io/go-nats"
)

var _ writers.Writer = (*exporter)(nil)

type exporter struct {
	writer writers.Writer
	logger log.Logger
}

const (
	// SenMLJSON represents SenML in JSON format content type.
	SenMLJSON = "application/senml+json"

	// SenMLCBOR represents SenML in CBOR format content type.
	SenMLCBOR = "application/senml+cbor"
)

var formats = map[string]senml.Format{
	SenMLJSON: senml.JSON,
	SenMLCBOR: senml.CBOR,
}

func New(nc *nats.Conn, repo writers.MessageRepository, transformer transformers.Transformer, channels map[string]bool, fConsume func(*nats.Msg), logger log.Logger) writers.Writer {
	e := exporter{logger: logger}
	w := writers.New(nc, repo, transformer, channels, e.Consume, logger)
	e.writer = w
	return &e
}

// Start method starts consuming messages received from NATS.
// This method transforms messages to SenML format before
// using MessageRepository to store them.
func (e *exporter) Start(queue string) error {
	return e.writer.Start(queue)
}

func (e *exporter) Consume(m *nats.Msg) {
	var msg mainflux.Message

	format, ok := formats[msg.ContentType]
	if !ok {
		format = senml.JSON
	}
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		e.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
		return
	}

	raw, err := senml.Decode(msg.Payload, format)
	if err != nil {
		e.logger.Error(fmt.Sprintf("Failed to decode payload message: %s", err))
	}
	msgs := []interface{}{}
	msgs = append(msgs, raw)

	e.Write(msgs)
}

func (e *exporter) Write(msgs ...interface{}) {
	e.writer.Write(msgs)
}
