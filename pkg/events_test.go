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
	"github.com/pebbe/zmq4"
	"testing"
	"time"
)

func TestEventOctopus_Configure(t *testing.T) {
	t.Run("creates ZMQ context", func(t *testing.T) {
		eo := &EventOctopus{}
		eo.Configure()

		if eo.zmqCtx == nil {
			t.Log("Expected zmqCtx to have been made")
		}
	})

	t.Run("creates ZMQ context", func(t *testing.T) {
		eo := &EventOctopus{}
		eo.Configure()

		if eo.feedbackChan == nil {
			t.Log("Expected feedback channel to have been made")
		}
	})
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

func TestEventOctopus_HappyFlow(t *testing.T) {
	s, _ := zmq4.NewSocket(zmq4.PUB)
	defer s.Close()
	s.Bind("ipc://bridge.ipc")

	eo := testEventOctopus()
	eo.Configure()
	eo.Start()
	defer eo.Shutdown()

	// todo
	time.Sleep(10 * time.Millisecond)

	t.Run("Event is received", func(t *testing.T) {
		event := fmt.Sprintf("random:state:id:produced")
		s.Send(event, 0)

		// todo
		time.Sleep(10 * time.Millisecond)

		e := eo.EventCallback.(*testEventCallback).event
		if e != event {
			t.Errorf("Expected to receive %s, got %s", event, e)
		}
	})
}

func testEventOctopus() *EventOctopus {
	eo := EventOctopus{
		Config:EventOctopusConfig{
			ZmqAddress:    "ipc://bridge.ipc",
			RestAddress:   ConfigRestAddressDefault,
			RetryInterval: 1,
		},
		EventCallback: &testEventCallback{},
	}

	return &eo
}

type testEventCallback struct {
	event string
}

func (t *testEventCallback) EventReceived(event *Event) {
	t.event = event.String()
}