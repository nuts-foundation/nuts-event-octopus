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
	core "github.com/nuts-foundation/nuts-go-core"
	"sync"
	"testing"
	"time"

	natsClient "github.com/nats-io/stan.go"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestEventOctopusInstance(t *testing.T) {
	i := EventOctopusInstance()

	t.Run("sets defaults", func(t *testing.T) {
		assert.Equal(t, i.Config.IncrementalBackoff, ConfigIncrementalBackoffDefault)
		assert.Equal(t, i.Config.MaxRetryCount, ConfigMaxRetryCountDefault)
		assert.Equal(t, i.Config.PurgeCompleted, false)
		assert.Equal(t, i.Config.AutoRecover, false)
		assert.Equal(t, i.Config.NatsPort, ConfigNatsPortDefault)
		assert.Equal(t, i.Config.Connectionstring, ConfigConnectionStringDefault)
	})
}

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

	t.Run("initializes the delayedConsumers", func(t *testing.T) {
		assert.NoError(t, eo.Start())

		assert.Equal(t, 5, len(eo.delayedConsumers))
	})
}

func TestEventOctopus_Shutdown(t *testing.T) {
	t.Run("Terminating does not give errors for default values", func(t *testing.T) {
		eo := testEventOctopus()
		_ = eo.configure()
		_ = eo.Shutdown()

		if err := eo.Shutdown(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func event() Event {
	return Event{
		RetryCount:           0,
		Payload:              "test",
		Name:                 EventConsentRequestConstructed,
		ExternalID:           "e_id",
		ConsentID:            uuid.NewV4().String(),
		InitiatorLegalEntity: "urn:nuts:entity:test",
	}
}

func TestEventOctopus_EventPersisted(t *testing.T) {
	i := testEventOctopus()
	defer i.Shutdown()
	i.Config.Connectionstring = "file::memory:?cache=shared&_busy_timeout=2500"
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

		e := event()
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

		e := event()
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

		e := event()
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

	t.Run("a published error event is stored in db", func(t *testing.T) {
		sc := stanConnection()
		defer sc.Close()
		emptyTable(i)

		u := uuid.NewV4().String()

		e := event()
		e.UUID = u
		e.Name = EventErrored

		je, _ := json.Marshal(e)
		_ = sc.Publish(ChannelConsentErrored, je)

		time.Sleep(100 * time.Millisecond)

		evt, err := i.GetEvent(u)
		if assert.NoError(t, err) {
			assert.NotNil(t, evt)
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
				event().Name: func(event *Event) {
					wg.Done()
				},
			})

		publisher, _ := i.EventPublisher("event-octopus-test")

		notify := make(chan bool)

		go func() {
			wg.Wait()
			notify <- true
		}()

		_ = publisher.Publish("EventRequestEvents", event())

		select {
		case <-notify:
			called = true
		case <-time.After(10 * time.Millisecond):
		}

		assert.True(t, called)
	})

	t.Run("unknown event is ignored", func(t *testing.T) {
		i := testEventOctopus()
		i.configure()
		i.Start()
		defer i.Shutdown()

		wg := sync.WaitGroup{}
		wg.Add(1)

		_ = i.Subscribe("event-logic",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				"other event": func(event *Event) {
					wg.Done()
				},
			})

		publisher, _ := i.EventPublisher("event-octopus-test")

		notify := make(chan bool)

		go func() {
			wg.Wait()
			notify <- true
		}()

		_ = publisher.Publish("EventRequestEvents", event())

		select {
		case <-notify:
			assert.Fail(t, "did not expect event to be handled")
		case <-time.After(10 * time.Millisecond):
		}
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
		_ = i.startStanServer() // use startStanServer() instead of Start() so there will not be a service for the event-store
		defer i.Shutdown()

		if len(i.stanClients) != 0 {
			t.Errorf("expected 0 clients, got %v", len(i.stanClients))
		}

		_ = i.Subscribe("event-logic",
			"EventRequestEvents",
			map[string]EventHandlerCallback{
				event().Name: func(event *Event) {},
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

func TestEventOctopus_Retry(t *testing.T) {
	i := testEventOctopus()
	_ = i.Configure()
	_ = i.Start()
	defer i.Shutdown()

	t.Run("event published to retry channel is picked up and republished to channel", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		uuid := uuid.NewV4().String()

		// the event
		var event = Event{
			RetryCount:           0,
			Payload:              "test",
			Name:                 EventConsentRequestConstructed,
			ExternalID:           "e_id",
			UUID:                 uuid,
			InitiatorLegalEntity: "urn:nuts:entity:test",
		}
		var receivedEvent *Event

		// subscription for waiting
		i.Subscribe(
			"test",
			ChannelConsentRequest,
			map[string]EventHandlerCallback{
				EventConsentRequestConstructed: func(event *Event) {
					receivedEvent = event
					wg.Done()
				},
			},
		)

		if err := i.publishEventToChannel(event, ChannelConsentRetry); err != nil {
			t.Fatal(err)
		}

		wg.Wait()

		assert.Equal(t, uuid, receivedEvent.UUID)
		assert.Equal(t, int(1), receivedEvent.RetryCount)
	})

	t.Run("event published to retry channel with max retry count is persisted as error", func(t *testing.T) {
		defer emptyTable(i)

		var e1 = Event{
			RetryCount:           ConfigMaxRetryCountDefault,
			Payload:              "test",
			Name:                 EventConsentRequestConstructed,
			ExternalID:           "e_id",
			UUID:                 uuid.NewV4().String(),
			InitiatorLegalEntity: "urn:nuts:entity:test",
		}

		if err := i.publishEventToChannel(e1, ChannelConsentRetry); err != nil {
			t.Fatal(err)
		}

		var e *Event
		poller := make(chan *Event)

		go func() {
			for {
				e, _ := i.GetEvent(e1.UUID)
				if e != nil {
					poller <- e
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()

		select {
		case e = <-poller:
		case <-time.After(50 * time.Millisecond):
		}

		if assert.NotNil(t, e) {
			assert.Equal(t, EventErrored, e.Name, e.String())
			if assert.NotNil(t, e.Error, e.String()) {
				assert.Equal(t, "max retry count reached", *e.Error)
			}
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
			RetryInterval:      ConfigRetryIntervalDefault,
			NatsPort:           ConfigNatsPortDefault,
			Connectionstring:   ConfigConnectionStringDefault,
			MaxRetryCount:      ConfigMaxRetryCountDefault,
			IncrementalBackoff: ConfigIncrementalBackoffDefault,
		},
		channelHandlers: make(map[string]map[string]ChannelHandlers),
		stanClients:     make(map[string]natsClient.Conn),
	}
}

func TestEventOctopus_Unsubscribe(t *testing.T) {
	i := testEventOctopus()
	i.startStanServer() // use startStanServer() instead of Start() so there will not be a service for the event-store
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

	t.Run("return error when subscription has been removed in lower level logic", func(t *testing.T) {
		i.Subscribe("service", "subject", map[string]EventHandlerCallback{"foo": func(event *Event) {}, "bar": func(event *Event) {}})

		// remove at stan level
		if assert.NoError(t, i.channelHandlers["service"]["subject"].subscription.Unsubscribe()) {
			assert.Error(t, i.Unsubscribe("service", "subject"))
		}
	})

}

func TestEventOctopus_Diagnostics(t *testing.T) {
	i := testEventOctopus()
	i.configure()
	i.Start()

	t.Run("Diagnostics returns 2 reports", func(t *testing.T) {
		results := i.Diagnostics()

		assert.Len(t, results, 2)
	})

	t.Run("Diagnostics returns Nats info", func(t *testing.T) {
		found := false
		results := i.Diagnostics()
		for _, r := range results {
			if r.Name() == "Nats streaming server" {
				found = true
				assert.Equal(t, "mode: STANDALONE @ 0.0.0.0:4222, ID: nuts, last error: NONE", r.String())
			}
		}

		assert.True(t, found)
	})

	t.Run("Diagnostics returns DB info", func(t *testing.T) {
		found := false
		results := i.Diagnostics()
		for _, r := range results {
			if r.Name() == "DB" {
				found = true
				assert.Equal(t, "ping: true", r.String())
			}
		}

		assert.True(t, found)
	})

	i.Shutdown()
	i.stanServer = nil

	t.Run("Diagnostics returns correct Nats info when down", func(t *testing.T) {
		found := false
		results := i.Diagnostics()
		for _, r := range results {
			if r.Name() == "Nats streaming server" {
				found = true
				assert.Equal(t, "DOWN", r.String())
			}
		}

		assert.True(t, found)
	})

	t.Run("Diagnostics returns DB info when down", func(t *testing.T) {
		found := false
		results := i.Diagnostics()
		for _, r := range results {
			if r.Name() == "DB" {
				found = true
				assert.Equal(t, "ping: false, error: sql: database is closed", r.String())
			}
		}

		assert.True(t, found)
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
	i.SaveOrUpdateEvent(e)

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
		i.SaveOrUpdateEvent(Event{
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

func TestEventOctopus_GetMode(t *testing.T) {
	assert.Equal(t, core.ServerEngineMode, EventOctopusConfig{}.GetMode())
}

func TestEventOctopus_purgeCompleted(t *testing.T) {
	t.Run("purgeCompleted removes events with name completed", func(t *testing.T) {
		i := testEventOctopus()
		i.configure()
		i.Start()
		defer i.Shutdown()

		// store new events
		i.SaveOrUpdateEvent(Event{
			ExternalID: "1",
			Name:       EventCompleted,
			UUID:       uuid.NewV4().String(),
		})

		i.SaveOrUpdateEvent(Event{
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
