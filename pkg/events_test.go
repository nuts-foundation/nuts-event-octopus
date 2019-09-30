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
	"encoding/json"
	"fmt"
	natsClient "github.com/nats-io/stan.go"
	uuid "github.com/satori/go.uuid"
	"sync"
	"testing"
	"time"
)

func TestEventOctopus_Configure(t *testing.T) {

}

func TestEventOctopus_Start(t *testing.T) {
	eo := testEventOctopus()
	_ = eo.Configure()
	defer eo.Shutdown()

	t.Run("feedbackChannel receives nil value on default config", func(t *testing.T) {
		if err := eo.Start(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestEventOctopus_Shutdown(t *testing.T) {
	t.Run("Terminating zmqCtx does not give errors for default values", func(t *testing.T) {
		eo := testEventOctopus()
		_ = eo.Configure()
		_ = eo.Shutdown()

		if err := eo.Shutdown(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

var event = Event{
	RetryCount:           0,
	Payload:              "test",
	Name:                 EventConsentRequestConstructed,
	ExternalId:           "e_id",
	ConsentId:            uuid.NewV4().String(),
	InitiatorLegalEntity: "urn:nuts:entity:test",
}

func TestEventOctopus_EventPersisted(t *testing.T) {
	i := testEventOctopus()
	defer i.Shutdown()
	i.Config.Connectionstring = "file::memory:?cache=shared"
	i.Configure()
	if err := i.Start(); err != nil {
		fmt.Printf("%v\n", err)
	}

	if err := i.RunMigrations(i.Db.DB()); err != nil {
		fmt.Printf("%v\n", err)
	}

	t.Run("a published event is persisted in db", func(t *testing.T) {
		stanClient := stanConnection()
		defer stanClient.Close()
		defer emptyTable(i)

		e := event
		u := uuid.NewV4().String()
		e.Uuid = u

		je, _ := json.Marshal(event)

		_ = stanClient.Publish(ChannelConsentRequest, je)

		time.Sleep(500 * time.Millisecond)

		evts, err := i.List()
		if err != nil {
			t.Error("Expected no error", err)
		}
		if len(*evts) != 1 {
			t.Errorf("Expected 1 event in DB, found %d", len(*evts))
		}
	})

	t.Run("a published event is updated in db", func(t *testing.T) {
		sc := stanConnection()
		defer sc.Close()
		defer emptyTable(i)

		u := uuid.NewV4().String()

		e := event
		e.Uuid = u

		je, _ := json.Marshal(e)
		_ = sc.Publish(ChannelConsentRequest, je)

		e.Name = EventConsentRequestInFlight

		je, _ = json.Marshal(e)
		_ = sc.Publish(ChannelConsentRequest, je)

		time.Sleep(100 * time.Millisecond)

		evts, _ := i.List()
		if len(*evts) != 1 {
			t.Fatalf("Expected to have received exactly 1 event, got %v", len(*evts))
		}
		if (*evts)[0].Name != EventConsentRequestInFlight {
			t.Errorf("Expected event to have name %s, found %s", EventConsentRequestInFlight, (*evts)[0].Name)
		}
	})
}

func TestEventOctopus_Subscribe(t *testing.T) {
	t.Run("subscribe to event with handler", func(t *testing.T) {
		called := false

		i := testEventOctopus()
		i.Configure()
		i.Start()
		defer i.Shutdown()

		_ = i.Subscribe("event-logic",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				event.Name: func(event *Event) {
					called = true
					t.Log("callback received")
				},
			})

		publisher, _ := i.EventPublisher("event-octopus-test")

		_ = publisher.Publish("EventRequestEvents", event)

		time.Sleep(500 * time.Millisecond)
		if !called {
			t.Error("callback should have been called")
		}
	})

	t.Run("adding handlers for the same service and subject should merge the handlers", func(t *testing.T) {
		i := testEventOctopus()
		i.Configure()
		i.Start()
		defer i.Shutdown()

		subject := "subject"
		service := "test"

		if _, ok := i.channelHandlers[service][subject]; ok {
			t.Error("expected an empty list fo channelhandlers")
		}

		_ = i.Subscribe(service,
			subject,
			map[string]EventHandlerCallback{
				"foo": func(event *Event) {},
			})

		if _, ok := i.channelHandlers[service][subject]; !ok {
			t.Error("expected a channelHandler")
		}

		_ = i.Subscribe(service,
			subject,
			map[string]EventHandlerCallback{
				"bar": func(event *Event) {},
			})

		if i.channelHandlers[service][subject].handlers["foo"] == nil {
			t.Error("expected handlers to be merged")
		}

		if i.channelHandlers[service][subject].handlers["bar"] == nil {
			t.Error("expected handlers to be merged")
		}
	})

	t.Run("two subscriptions for the same service should result in one connection", func(t *testing.T) {
		i := testEventOctopus()
		_ = i.nats() // use nats() instead of Start() so there will not be a service for the event-store
		defer i.Shutdown()

		if len(i.stanClients) != 0 {
			t.Errorf("expected 0 clients, got %v", len(i.stanClients))
		}

		_ = i.Subscribe("event-logic",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				event.Name: func(event *Event) {},
			})

		if len(i.stanClients) != 1 {
			t.Errorf("expected 1 clients, got %v", len(i.stanClients))
		}

		_ = i.Subscribe("event-logic",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				"foo": func(event *Event) {},
			})

		if len(i.stanClients) != 1 {
			t.Errorf("expected 1 clients, got %v", len(i.stanClients))
		}

		_ = i.Subscribe("other-service",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				"bar": func(event *Event) {},
			})

		if len(i.stanClients) != 2 {
			t.Errorf("expected 2 clients, got %v", len(i.stanClients))
		}

	})
}

func emptyTable(eo *EventOctopus) {
	event := &Event{}
	defer eo.Db.Delete(&event)
}

func stanConnection() natsClient.Conn {
	sc, err := natsClient.Connect(
		"nuts",
		"event-octopus-test",
		natsClient.NatsURL("nats://localhost:4222"),
	)

	if err != nil {
		panic(err)
	}

	return sc
}

func testEventOctopus() *EventOctopus {
	return EventOctopusInstance()
}

func TestEventOctopus_Unsubscribe(t *testing.T) {
	i := testEventOctopus()
	i.nats() // use nats() instead of Start() so there will not be a service for the event-store
	defer i.Shutdown()
	i.Subscribe("service", "subject", map[string]EventHandlerCallback{"foo": func(event *Event) {}, "bar": func(event *Event) {}})
	i.Subscribe("service", "other-subject", map[string]EventHandlerCallback{"foo": func(event *Event) {}, "bar": func(event *Event) {}})

	if len(i.channelHandlers["service"]) != 2 {
		t.Errorf("expected 2 channelHandlers, got %v", len(i.channelHandlers["service"]))
	}

	if len(i.stanClients) != 1 {
		t.Error("expected 1 connections")
	}

	i.Unsubscribe("service", "subject")

	if len(i.channelHandlers["service"]) != 1 {
		t.Errorf("expected 1 channelHandlers, got %v", len(i.channelHandlers["service"]))
	}

	if len(i.stanClients) != 1 {
		t.Error("expected 1 connections")
	}

	i.Unsubscribe("service", "other-subject")

	if len(i.channelHandlers["service"]) != 0 {
		t.Errorf("expected 0 channelHandlers, got %v", len(i.channelHandlers["service"]))
	}

	if len(i.stanClients) != 0 {
		t.Error("expected all connections to be closed")
	}
}

func TestEventOctopus_recover(t *testing.T) {
	t.Run("events not completed are published", func(t *testing.T) {
		i := testEventOctopus()
		i.Configure()
		i.Start()
		defer i.Shutdown()

		sc := stanConnection()
		defer sc.Close()

		event := &Event{}

		// test synchronization
		wg := sync.WaitGroup{}
		wg.Add(1)

		// subscribe to Nats
		sc.Subscribe(ChannelConsentRequest, func(msg *natsClient.Msg) {
			// Unmarshal JSON that represents the Order data
			json.Unmarshal(msg.Data, &event)
			wg.Done()
		})

		// store new event
		i.SaveOrUpdate(Event{
			ExternalId: "2",
			Name:       EventConsentDistributed,
			Uuid:       uuid.NewV4().String(),
		})

		// start recover
		i.recover()

		// wait for event
		wg.Wait()

		if event.ExternalId != "2" {
			t.Error("Expected to receive event with externalId 2")
		}
	})

}

func TestEventOctopus_purgeCompleted(t *testing.T) {
	t.Run("purgeCompleted removes events with name completed", func(t *testing.T) {
		i := testEventOctopus()
		i.Configure()
		i.Start()
		defer i.Shutdown()

		// store new events
		i.SaveOrUpdate(Event{
			ExternalId: "1",
			Name:       EventCompleted,
			Uuid:       uuid.NewV4().String(),
		})

		i.SaveOrUpdate(Event{
			ExternalId: "2",
			Name:       EventConsentDistributed,
			Uuid:       uuid.NewV4().String(),
		})

		i.purgeCompleted()

		e, _ := i.GetEventByExternalId("1")

		if e != nil {
			t.Error("Expected event with externalId 1 to be deleted")
		}

		e2, _ := i.GetEventByExternalId("2")

		if e2 == nil {
			t.Error("Expected event with externalId 2 not to be deleted")
		}
	})
}
