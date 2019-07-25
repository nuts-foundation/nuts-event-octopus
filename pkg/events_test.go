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
	"testing"
	"time"
)

func TestEventOctopus_Configure(t *testing.T) {

}

func TestEventOctopus_Start(t *testing.T) {
	eo := testEventOctopus()
	eo.Configure()
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
		eo.Configure()
		eo.Shutdown()

		if err := eo.Shutdown(); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestEventOctopus_EventPersisted(t *testing.T) {
	i := EventOctopusIntance()
	i.Config.Connectionstring = "file::memory:?cache=shared"
	if err := i.Start(); err != nil {
		fmt.Printf("%v\n", err)
	}

	if err := i.RunMigrations(i.Db.DB()); err != nil {
		fmt.Printf("%v\n", err)
	}

	event := Event{
		RetryCount: 0,
		Payload:    "test",
		Name:       EventStateOffered,
		ExternalId: "e_id",
		ConsentId:  uuid.NewV4().String(),
		Custodian:  "urn:nuts:custodian:test",
	}

	t.Run("a published event is persisted in db", func(t *testing.T) {
		sc := stanConnection()
		defer sc.Close()
		defer emptyTable()

		u := uuid.NewV4().String()

		e := event
		e.Uuid = u

		je, _ := json.Marshal(event)

		sc.Publish(ChannelConsentRequest, je)

		time.Sleep(30 * time.Millisecond)

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
		defer emptyTable()

		u := uuid.NewV4().String()

		e := event
		e.Uuid = u

		je, _ := json.Marshal(e)
		sc.Publish(ChannelConsentRequest, je)

		e.Name = EventStateCompleted

		je, _ = json.Marshal(e)
		sc.Publish(ChannelConsentRequest, je)

		time.Sleep(100 * time.Millisecond)

		evts, _ := i.List()
		if len(*evts) != 1 {
			t.Fatalf("Expected to have received exactly 1 event, got %v", len(*evts))
		}
		if (*evts)[0].Name != EventStateCompleted {
			t.Errorf("Expected event to have name %s, found %s", EventStateCompleted, (*evts)[0].Name)
		}
	})
}

func emptyTable() {
	event := &Event{}
	i := EventOctopusIntance()
	defer i.Db.Delete(&event)
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
	eo := EventOctopus{
		Config: EventOctopusConfig{
			RetryInterval: 1,
			NatsPort:      4222,
		},
	}

	return &eo
}
