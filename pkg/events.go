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
	"context"
	"fmt"
	bridgeClient "github.com/nuts-foundation/consent-bridge-go-client/api"
	zmq "github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const ConfigEpoch = "eventStartEpoch"
const ConfigZmqAddress = "ZmqAddress"
const ConfigRestAddress = "RestAddress"
const ConfigRetryInterval = "retryInterval"

const ConfigEpochDefault = 0
const ConfigZmqAddressDefault = "tcp://127.0.0.1:5563"
const ConfigRestAddressDefault = "http://localhost:8080"
const ConfigRetryIntervalDefault = 60

type EventOctopusConfig struct {
	EventStartEpoch int
	ZmqAddress      string
	RestAddress     string
	RetryInterval   int
}

const (
	Produced = "produced"
	Consumed = "consumed"
)

type Event struct {
	// Topic with which messages are filtered
	Topic  string

	// State is the subject of the event
	State  string

	// Id represents the externalId of that changed state
	Id     string
	// Action: consumed or produced
	Action string
}

func (e *Event) fromString(s string) {
	splitted := strings.Split(s, ":")
	e.Topic = splitted[0]
	e.State = splitted[1]
	e.Id = splitted[2]
	e.Action = splitted[3]
}

func (e *Event) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", e.Topic, e.State, e.Id, e.Action)
}

type EventCallback interface {
	EventReceived(event *Event)
}

// default implementation for EventOctopusInstance
type EventOctopus struct {
	Config        EventOctopusConfig
	configOnce    sync.Once
	configDone    bool
	zmqCtx	      *zmq.Context
	feedbackChan  chan error
	EventCallback EventCallback
	shutdown      bool
}

var instance *EventOctopus
var oneInstance sync.Once

type dummy struct{}
func (*dummy) EventReceived(event *Event) {
	logrus.Infof("receieved %v", event)
}

func EventOctopusIntance() *EventOctopus {
	oneInstance.Do(func() {
		instance = &EventOctopus{
			Config: EventOctopusConfig{
				RetryInterval: 60,
			},
			EventCallback: &dummy{},
		}
	})
	return instance
}

func (octopus *EventOctopus) bridgeEventListener() {
	//  Socket to talk to dispatcher
	s, _ := zmq.NewSocket(zmq.SUB)
	defer s.Close()

	rnd := "random"
	err := s.Connect(octopus.Config.ZmqAddress)
	if err != nil {
		octopus.feedbackChan <- fmt.Errorf("connecting to bridge: %v", err)
		return
	}

	logrus.Infof("Connected stream to Nuts consent bridge @ %s", octopus.Config.ZmqAddress)

	err = s.SetSubscribe(rnd)
	if err != nil {
		octopus.feedbackChan <- fmt.Errorf("subscribing to bridge: %v", err)
		return
	}
	logrus.Infof("Set subscription filter to [%s]", rnd)

	// socket ready, no config problems
	octopus.feedbackChan <- nil

	// send start msg
	go octopus.initStreamWithRetry(rnd)

	for {
		msg, err := s.Recv(0)
		if err != nil {
			logrus.Errorf("Received error %v on s.Recv", err)
			break
		}

		logrus.Debugf("Received %v", msg)
		if octopus.EventCallback != nil {
			e := &Event{}
			e.fromString(msg)
			octopus.EventCallback.EventReceived(e)
		}
		// call nuts-consent-logic
	}
}

func (octopus *EventOctopus) initStreamWithRetry(rnd string) {
	// todo http/https scheme config
	// send start message
	client := bridgeClient.NewConsentBridgeClient()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err := client.InitEventStream(ctx, bridgeClient.EventStreamSetting{
		Epoch: int64(octopus.Config.EventStartEpoch),
		Topic: rnd,
	})

	if err != nil {
		logrus.Warnf("Stream not ready: %v", err)

		if !octopus.shutdown {
			logrus.Infof("Retrying in %d seconds", octopus.Config.RetryInterval)

			time.Sleep(time.Duration(octopus.Config.RetryInterval) * time.Second)

			octopus.initStreamWithRetry(rnd)
		} else {
			logrus.Infof("Shutdown called, stopping retry")
		}
	}
}

// Configure initiates a ZQM context
func (octopus *EventOctopus) Configure() error {
	var err error

	octopus.zmqCtx, err = zmq.NewContext()
	octopus.feedbackChan = make(chan error)

	return err
}

// Start starts the receiver socket in a go routine
func (octopus *EventOctopus) Start() error {
	go octopus.bridgeEventListener()

	return <- octopus.feedbackChan
}

// Shutdown closes the ZMQ context which basically shutsdown all ZMQ sockets
func (octopus *EventOctopus) Shutdown() error {
	var err error

	octopus.shutdown = true

	if octopus.zmqCtx != nil {
		err = octopus.zmqCtx.Term()
	}

	return err
}
