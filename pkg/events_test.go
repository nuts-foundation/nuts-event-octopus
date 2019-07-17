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
	//i.Config.Connectionstring = "test.db"
	i.Start()
	i.RunMigrations(i.Db.DB())

	t.Run("a published event is persisted in db", func(t *testing.T) {
		sc := stanConnection()

		u := uuid.NewV4().String()

		event := Event{
			Uuid: u,
			RetryCount: 0,
			Payload: "test",
			State: "offered",
			ExternalId: "e_id",
			ConsentId: uuid.NewV4().String(),
			Custodian: "urn:nuts:custodian:test",
		}

		je, _ := json.Marshal(event)

		sc.Publish("consent-request", je)

		time.Sleep(time.Second)

		evts, _ := i.List()
		if len(*evts) != 1 {
			t.Errorf("Expected 1 event in DB, found %d", len(*evts))
		}
	})
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
		},
	}

	return &eo
}
