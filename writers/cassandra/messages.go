// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"errors"

	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

var _ writers.MessageRepository = (*cassandraRepository)(nil)

type cassandraRepository struct {
	session *gocql.Session
}

// New instantiates Cassandra message repository.
func New(session *gocql.Session) writers.MessageRepository {
	return &cassandraRepository{session}
}

func (cr *cassandraRepository) Save(messages ...interface{}) error {
	cql := `INSERT INTO messages (id, channel, subtopic, publisher, protocol,
			name, unit, value, string_value, bool_value, data_value, sum,
			time, update_time, link)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id := gocql.TimeUUID()

	for _, m := range messages {
		msg, ok := m.(senml.Message)
		if !ok {
			return errors.New("incorrect message type")
		}
		err := cr.session.Query(cql, id, msg.Channel, msg.Subtopic, msg.Publisher,
			msg.Protocol, msg.Name, msg.Unit, msg.Value, msg.StringValue,
			msg.BoolValue, msg.DataValue, msg.Sum, msg.Time, msg.UpdateTime, msg.Link).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
