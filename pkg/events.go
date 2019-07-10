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
	"github.com/nuts-foundation/nuts-consent-logic/pkg"
	zmq "github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const ConfigEpoch = "eventStartEpoch"
const ConfigZmqAddress = "ZmqAddress"
const ConfigRetryInterval = "retryInterval"

const ConfigEpochDefault = 0
const ConfigZmqAddressDefault = "tcp://127.0.0.1:5563"
const ConfigRetryIntervalDefault = 60

type EventOctopusConfig struct {
	EventStartEpoch int
	ZmqAddress      string
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
	Id     LinearId
	// Action: consumed or produced
	Action string
}

// LinearId represents the combined id of externalId and UUID
type LinearId struct {
	ExternalId 	*string
	UUID 		string
}

func LinearIdFromString(s string) LinearId {
	if strings.Contains(s, "_") {
		splitted := strings.Split(s, "_")
		return LinearId{
			ExternalId: &splitted[0],
			UUID:       splitted[1],
		}
	} else {
		return LinearId{
			UUID:       s,
		}
	}
}

func (e *Event) fromString(s string) {
	splitted := strings.Split(s, ":")
	e.Topic = splitted[0]
	e.State = splitted[1]
	e.Id = LinearIdFromString(splitted[2])
	e.Action = splitted[3]
}

func (e *Event) String() string {
	return fmt.Sprintf("%s:%s:%v:%s", e.Topic, e.State, e.Id, e.Action)
}

func (l LinearId) String() string {
	if l.ExternalId != nil {
		return fmt.Sprintf("%s_%s", *l.ExternalId, l.UUID)
	} else {
		return fmt.Sprintf("%s", l.UUID)
	}
}

type EventCallback interface {
	EventReceived(event *Event)
}

// default implementation for EventOctopusInstance
type EventOctopus struct {
	Config        EventOctopusConfig
	configOnce    sync.Once
	configDone    bool
	zmqCtx        *zmq.Context
	feedbackChan  chan error
	eventCallback EventCallback
	shutdown      bool
}

var instance *EventOctopus
var oneInstance sync.Once

type ConsentRequestCallback struct{
	consentLogic  pkg.ConsentLogicClient
}

func (crc *ConsentRequestCallback) EventReceived(event *Event) {
	logrus.Debugf("received %v", event)

	if err := crc.consentLogic.HandleConsentRequest(event.Id.UUID); err != nil {
		// todo serious error, missing events? => persist and retry!
		logrus.Errorf("Error in sending event to eventLogic: %v", err)
	}
}

// EventOctopusIntance returns the EventOctopus singleton
func EventOctopusIntance() *EventOctopus {
	oneInstance.Do(func() {
		instance = &EventOctopus{
			Config: EventOctopusConfig{
				RetryInterval: 60,
			},
			eventCallback: &ConsentRequestCallback{
				consentLogic: pkg.NewConsentLogicClient(),
			},
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
		if octopus.eventCallback != nil {
			e := &Event{}
			e.fromString(msg)
			octopus.eventCallback.EventReceived(e)
		}
		// call nuts-consent-logic
	}
}

func (octopus *EventOctopus) initStreamWithRetry(rnd string) {
	// todo http/https scheme config
	// send start message
	client := bridgeClient.NewConsentBridgeClient()
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 10)
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
