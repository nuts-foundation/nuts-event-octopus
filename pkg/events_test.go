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
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestEventOctopus_Configure(t *testing.T) {
	t.Run("Configure runs only once", func(t *testing.T) {
		i := testEventOctopus()
		i.Configure()
		var ranTwice bool

		i.configOnce.Do(func() {
			ranTwice = true
		})

		assert.False(t, ranTwice)
	})
}

func TestEventOctopus_Start(t *testing.T) {
	eo := testEventOctopus()
	_ = eo.configure()
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
		_ = eo.configure()
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
	ExternalID:           "e_id",
	ConsentID:            uuid.NewV4().String(),
	InitiatorLegalEntity: "urn:nuts:entity:test",
}

func TestEventOctopus_EventPersisted(t *testing.T) {
	i := testEventOctopus()
	defer i.Shutdown()
	i.Config.Connectionstring = "file::memory:?cache=shared"
	i.configure()
	if err := i.Start(); err != nil {
		fmt.Printf("%v\n", err)
	}

	if err := i.RunMigrations(i.Db.DB()); err != nil {
		fmt.Printf("%v\n", err)
	}

	t.Run("a published event is persisted in db", func(t *testing.T) {
		stanClient := stanConnection()
		defer stanClient.Close()
		emptyTable(i)

		e := event
		u := uuid.NewV4().String()
		e.UUID = u

		je, _ := json.Marshal(event)

		_ = stanClient.Publish(ChannelConsentRequest, je)

		time.Sleep(500 * time.Millisecond)

		evts, err := i.List()
		if assert.Nil(t, err) {
			assert.Equal(t, 1, len(*evts))
		}
	})

	t.Run("an incorrect event is persisted in db as errored", func(t *testing.T) {
		stanClient := stanConnection()
		defer stanClient.Close()
		emptyTable(i)

		e := event
		u := uuid.NewV4().String()
		e.UUID = u

		je := []byte("{")


		_ = stanClient.Publish(ChannelConsentRequest, je)

		time.Sleep(500 * time.Millisecond)

		evts, err := i.List()

		if assert.Nil(t, err) {
			assert.Equal(t, 1, len(*evts))
			assert.Equal(t, EventErrored, (*evts)[0].Name)
		}
	})

	t.Run("a published event is updated in db", func(t *testing.T) {
		sc := stanConnection()
		defer sc.Close()
		emptyTable(i)

		u := uuid.NewV4().String()

		e := event
		e.UUID = u

		je, _ := json.Marshal(e)
		_ = sc.Publish(ChannelConsentRequest, je)

		e.Name = EventConsentRequestInFlight

		je, _ = json.Marshal(e)
		_ = sc.Publish(ChannelConsentRequest, je)

		time.Sleep(100 * time.Millisecond)

		evts, err := i.List()
		if assert.Nil(t, err) {
			assert.Equal(t, 1, len(*evts))
			assert.Equal(t, EventConsentRequestInFlight, (*evts)[0].Name)
		}
	})
}

func TestEventOctopus_EventPublisher(t *testing.T) {
	t.Run("Returns error when not initialized", func(t *testing.T) {
		i := testEventOctopus()
		_, err := i.EventPublisher("ID")

		assert.NotNil(t, err)
	})
}

func TestEventOctopus_Subscribe(t *testing.T) {
	t.Run("subscribe to event with handler", func(t *testing.T) {
		called := false

		i := testEventOctopus()
		i.configure()
		i.Start()
		defer i.Shutdown()

		wg := sync.WaitGroup{}
		wg.Add(1)

		_ = i.Subscribe("event-logic",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				event.Name: func(event *Event) {
					wg.Done()
				},
			})

		publisher, _ := i.EventPublisher("event-octopus-test")

		notify := make(chan bool)

		go func() {
			wg.Wait()
			notify <- true
		}()

		_ = publisher.Publish("EventRequestEvents", event)

		select {
			case <-notify:called = true
			case <-time.After(10 * time.Millisecond):
		}

		assert.True(t, called)
	})

	t.Run("adding handlers for the same service and subject should merge the handlers", func(t *testing.T) {
		i := testEventOctopus()
		i.configure()
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
	eo.Db.Delete(&event)
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
	return &EventOctopus{
		Name: Name,
		Config: EventOctopusConfig{
			RetryInterval:    ConfigRetryIntervalDefault,
			NatsPort:         ConfigNatsPortDefault,
			Connectionstring: ConfigConnectionStringDefault,
		},
		channelHandlers: make(map[string]map[string]ChannelHandlers),
		stanClients:     make(map[string]natsClient.Conn),
	}
}

func TestEventOctopus_Unsubscribe(t *testing.T) {
	i := testEventOctopus()
	i.nats() // use nats() instead of Start() so there will not be a service for the event-store
	defer i.Shutdown()

	t.Run("Subscribe and unsubscribe count", func(t *testing.T) {
		i.Subscribe("service", "subject", map[string]EventHandlerCallback{"foo": func(event *Event) {}, "bar": func(event *Event) {}})
		i.Subscribe("service", "other-subject", map[string]EventHandlerCallback{"foo": func(event *Event) {}, "bar": func(event *Event) {}})

		assert.Equal(t, 2, len(i.channelHandlers["service"]))
		assert.Equal(t, 1, len(i.stanClients))

		i.Unsubscribe("service", "subject")

		assert.Equal(t, 1, len(i.channelHandlers["service"]))
		assert.Equal(t, 1, len(i.stanClients))

		i.Unsubscribe("service", "other-subject")

		assert.Equal(t, 0, len(i.channelHandlers["service"]))
		assert.Equal(t, 0, len(i.stanClients))
	})

	t.Run("Unknown subscription returns error", func(t *testing.T) {
		err := i.Unsubscribe("unknown", "unknown")

		assert.NotNil(t, err)
	})

}

func TestEventOctopus_GetEvent(t *testing.T) {
	i := testEventOctopus()
	i.configure()
	i.Start()
	defer i.Shutdown()

	// store new event
	e := Event{
		ExternalID: "2",
		Name:       EventConsentDistributed,
		UUID:       uuid.NewV4().String(),
	}
	i.SaveOrUpdate(e)

	t.Run("event found", func(t *testing.T) {
		ep, err := i.GetEvent(e.UUID)
		if assert.Nil(t, err) {
			assert.Equal(t, EventConsentDistributed, ep.Name)
		}
	})

	t.Run("event not found", func(t *testing.T) {
		ep, err := i.GetEvent(uuid.NewV4().String())
		assert.Nil(t, err)
		assert.Nil(t, ep)
	})
}

func TestEventOctopus_recover(t *testing.T) {
	t.Run("events not completed are published", func(t *testing.T) {
		i := testEventOctopus()
		i.configure()
		i.Start()

		sc := stanConnection()
		defer sc.Close()
		defer i.Shutdown()

		event := &Event{}

		// store new event
		i.SaveOrUpdate(Event{
			ExternalID: "2",
			Name:       EventConsentDistributed,
			UUID:       uuid.NewV4().String(),
		})

		// test synchronization
		wg := sync.WaitGroup{}
		events, _ := i.allEvents()
		wg.Add(len(events))

		// subscribe to Nats
		sc.Subscribe(ChannelConsentRequest, func(msg *natsClient.Msg) {
			// Unmarshal JSON that represents the Order data
			json.Unmarshal(msg.Data, &event)
			wg.Done()
		})

		// start recover
		i.recover()

		// wait for event
		wg.Wait()

		if event.ExternalID != "2" {
			t.Error("Expected to receive event with externalId 2")
		}

		// otherwise ack happens when disconnecting
		time.Sleep(10 * time.Millisecond)
	})

}

func TestEventOctopus_purgeCompleted(t *testing.T) {
	t.Run("purgeCompleted removes events with name completed", func(t *testing.T) {
		i := testEventOctopus()
		i.configure()
		i.Start()
		defer i.Shutdown()

		// store new events
		i.SaveOrUpdate(Event{
			ExternalID: "1",
			Name:       EventCompleted,
			UUID:       uuid.NewV4().String(),
		})

		i.SaveOrUpdate(Event{
			ExternalID: "2",
			Name:       EventConsentDistributed,
			UUID:       uuid.NewV4().String(),
		})

		i.purgeCompleted()

		e, _ := i.GetEventByExternalID("1")

		if e != nil {
			t.Error("Expected event with externalId 1 to be deleted")
		}

		e2, _ := i.GetEventByExternalID("2")

		if e2 == nil {
			t.Error("Expected event with externalId 2 not to be deleted")
		}
	})
}
