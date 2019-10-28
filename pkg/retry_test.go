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
	"github.com/nats-io/stan.go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDelayedConsumer(t *testing.T) {
	i := testEventOctopus()
	_ = i.nats()
	defer i.Shutdown()

	t.Run("delayed event is received on correct publish channel", func(t *testing.T) {
		sc := conn("ok")
		defer sc.Close()

		dc := DelayedConsumer{
			consumeSubject: "channelIn",
			publishSubject: "channelOut",
			conn:           sc,
			delay:          10 * time.Millisecond,
		}

		if assert.Nil(t, dc.Start()) {
			found := false

			sc.Subscribe("channelOut", func(msg *stan.Msg) {
				found = true
			})

			sc.Publish("channelIn", []byte("test"))

			assert.False(t, found)
			time.Sleep(15 * time.Millisecond)
			assert.True(t, found)
		}
	})

	t.Run("when not connected for publishing, msg remains", func(t *testing.T) {
		sc := conn("unconsume")

		dc := DelayedConsumer{
			consumeSubject: "channelIn",
			publishSubject: "channelOut",
			conn:           sc,
			delay:          10 * time.Millisecond,
		}

		if assert.Nil(t, dc.Start()) {
			pendingBefore, _, _ := dc.subscription.Pending()

			sc.Publish("channelIn", []byte("test"))

			sc.Close()
			time.Sleep(15 * time.Millisecond)

			// restart subscription
			dc.conn = conn("unconsume")
			assert.Nil(t, dc.Start())
			pendingAfter, _, _ := dc.subscription.Pending()

			assert.Equal(t, 1, pendingAfter-pendingBefore)
		}
	})
}

func TestDelayedConsumer_Stop(t *testing.T) {
	i := testEventOctopus()
	_ = i.nats()
	defer i.Shutdown()

	t.Run("stop closes the subscription", func(t *testing.T) {
		sc := conn("close")

		dc := DelayedConsumer{
			consumeSubject: "channelIn",
			publishSubject: "channelOut",
			conn:           sc,
			delay:          10 * time.Millisecond,
		}

		if assert.Nil(t, dc.Start()) {
			assert.Nil(t, dc.Stop())
			assert.False(t, dc.subscription.IsValid())
		}
	})
}

func TestNewDelayedConsumerSet(t *testing.T) {
	set := NewDelayedConsumerSet("in", "out", 3, time.Millisecond, 2, nil)

	t.Run("gives the correct number of DelayedConsumer", func(t *testing.T) {
		assert.Equal(t, 3, len(set))
	})

	t.Run("sets correct in-channels", func(t *testing.T) {
		assert.Equal(t, "in-0", set[0].consumeSubject)
		assert.Equal(t, "in-1", set[1].consumeSubject)
		assert.Equal(t, "in-2", set[2].consumeSubject)
	})

	t.Run("sets correct out-channels", func(t *testing.T) {
		assert.Equal(t, "out", set[0].publishSubject)
		assert.Equal(t, "out", set[1].publishSubject)
		assert.Equal(t, "out", set[2].publishSubject)
	})

	t.Run("sets correct delays", func(t *testing.T) {
		assert.Equal(t, time.Millisecond, set[0].delay)
		assert.Equal(t, 2*time.Millisecond, set[1].delay)
		assert.Equal(t, 4*time.Millisecond, set[2].delay)
	})
}

func conn(id string) stan.Conn {
	sc, err := stan.Connect(
		"nuts",
		id,
		stan.NatsURL("nats://localhost:4222"),
	)

	if err != nil {
		panic(err)
	}

	return sc
}
