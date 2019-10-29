/*
 * Nuts event octopus
 * Copyright (C) 2019. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package pkg

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"time"
)

// DelayedConsumer holds info for creating a subscription on Nats for consuming events and re-publishing them with a certain delay
type DelayedConsumer struct {
	consumeSubject string        // Channel/topic to read from
	publishSubject string        // Channel to publish to
	delay          time.Duration // time to wait for sending ack
	conn           stan.Conn     // ackWait must match!
	subscription   stan.Subscription
	shutdown 	   bool
}

// Start starts the subscription on the given connection
func (dc *DelayedConsumer) Start() error {
	var err error

	dc.subscription, err = dc.conn.Subscribe(dc.consumeSubject, func(msg *stan.Msg) {
		// delegate to go procedure
		go dc.delayedPublishAndAck(msg)
	}, stan.DurableName(fmt.Sprintf("%s-%s", dc.consumeSubject, "durable")),
		stan.AckWait(time.Second+dc.delay), // some extra time for publishing
		stan.SetManualAckMode(),
		stan.StartWithLastReceived(),
	)

	if err != nil {
		return err
	}

	logrus.Debugf("started delayed consumer for subject: %s", dc.consumeSubject)

	return nil
}

func (dc *DelayedConsumer) delayedPublishAndAck(msg *stan.Msg) {
	targetTime := time.Now().Add(dc.delay)

	for {
		time.Sleep(10 * time.Millisecond)
		if dc.shutdown {
			return
		}
		if time.Now().After(targetTime) {
			break
		}
	}

	if dc.conn.NatsConn() != nil {
		if dc.conn.NatsConn().IsConnected() {
			if err := dc.conn.Publish(dc.publishSubject, msg.Data); err != nil {
				logrus.WithError(err).Fatal("failed to publish delayed message")
			}
			if err := msg.Ack(); err != nil {
				logrus.WithError(err).Fatal("failed to ack retry message")
			}
		} else {
			logrus.Warnf("ignoring retry message, no connection available, current status: %d", dc.conn.NatsConn().Status())
		}
	}
}

// NewDelayedConsumerSet creates a set of DelayedConsumer where each successive poller has a interval which is exponent times bigger than the previous one
func NewDelayedConsumerSet(consumeSubject string, publishSubject string, count int, interval time.Duration, exponent int, conn stan.Conn) []*DelayedConsumer {
	var pollers []*DelayedConsumer

	expInterval := interval
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%d", consumeSubject, i)

		pollers = append(pollers, &DelayedConsumer{
			consumeSubject: name,
			publishSubject: publishSubject,
			delay:          expInterval,
			conn:           conn,
		})

		expInterval *= time.Duration(exponent)
	}

	return pollers
}

// Stop stops the consumer
func (dc *DelayedConsumer) Stop() error {
	dc.shutdown = true
	return dc.subscription.Close()
}
