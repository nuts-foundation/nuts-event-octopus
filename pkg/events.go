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
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jinzhu/gorm"
	natsServer "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats-streaming-server/stores"
	natsClient "github.com/nats-io/stan.go"
	"github.com/nuts-foundation/nuts-event-octopus/migrations"
	"github.com/sirupsen/logrus"
	"sync"
)

const ConfigRetryInterval = "retryInterval"
const ConfigNatsPort = "natsPort"
const ConfigConnectionstring = "connectionstring"

const ConfigRetryIntervalDefault = 60
const ConfigNatsPortDefault = 4222
const ConfigConnectionStringDefault = "file:not_used?mode=memory&cache=shared"

// EventOctopusConfig holds the config for the EventOctopusInstance
type EventOctopusConfig struct {
	RetryInterval    int
	NatsPort         int
	Connectionstring string
}

// EventOctopusClient is the client interface for publishing events
type EventOctopusClient interface {
	EventPublisher(clientId string) (natsClient.Conn, error)
	Subscribe(service, subject string, callbacks map[string]EventHandlerCallback) error
}

type ChannelHandlers struct {
	subscription natsClient.Subscription
	handlers     map[string]EventHandlerCallback
}

// EventOctopus is the default implementation for EventOctopusInstance
type EventOctopus struct {
	Config     EventOctopusConfig
	configOnce sync.Once
	configDone bool
	stanServer *natsServer.StanServer
	Db         *gorm.DB
	// Clients per service
	stanClients     map[string]natsClient.Conn
	channelHandlers map[string]map[string]ChannelHandlers
}

var instance *EventOctopus
var oneInstance sync.Once

// EventOctopusIntance returns the EventOctopus singleton
func EventOctopusIntance() *EventOctopus {
	oneInstance.Do(func() {
		instance = &EventOctopus{
			Config: EventOctopusConfig{
				RetryInterval:    ConfigRetryIntervalDefault,
				NatsPort:         ConfigNatsPortDefault,
				Connectionstring: ConfigConnectionStringDefault,
			},
			channelHandlers: make(map[string]map[string]ChannelHandlers),
			stanClients:     make(map[string]natsClient.Conn),
		}
	})
	return instance
}

// Subscribe lets you subscribe to events for a service and subject. For each Event.name you can provide a callback function
func (octopus *EventOctopus) Subscribe(service, subject string, handlers map[string]EventHandlerCallback) error {
	// create a new ChannelHandler if it does not exists for the combination of service and subject
	if channelHandlers, ok := octopus.channelHandlers[service][subject]; !ok {

		channelHandlers := ChannelHandlers{
			handlers: handlers,
		}
		var stanClient natsClient.Conn
		if stanClient, ok = octopus.stanClients[service]; !ok {
			var err error
			if stanClient, err = octopus.Client(service); err != nil {
				return err
			}
			octopus.stanClients[service] = stanClient
		}

		channelHandlers.subscription, _ = stanClient.Subscribe(subject, func(msg *natsClient.Msg) {
			event := &Event{}
			// Unmarshal JSON that represents the Order data
			err := json.Unmarshal(msg.Data, &event)
			if err != nil {
				logrus.Errorf("Error unmarshalling event: %v", err)
				return
			}
			handler := channelHandlers.handlers[event.Name]
			if handler == nil {
				logrus.Infof("Event without handler %v", event.Name)
				return
			}
			handler(event)
		})
		// does the inner map exists?
		if _, ok := octopus.channelHandlers[service]; !ok {
			octopus.channelHandlers[service] = make(map[string]ChannelHandlers)
		}
		octopus.channelHandlers[service][subject] = channelHandlers
	} else {
		// merge handlers
		for key, handler := range handlers {
			channelHandlers.handlers[key] = handler
		}
	}
	return nil
}

func (octopus *EventOctopus) Unsubscribe(service, subject string) error {
	handlers, ok := octopus.channelHandlers[service][subject]
	if !ok {
		return fmt.Errorf("no subscription found for %s.%s", service, subject)
	}

	if err := handlers.subscription.Unsubscribe(); err != nil {
		return err
	}
	// delete subject from channelHandlers
	delete(octopus.channelHandlers[service], subject)

	// if this was the only subject for this service, remove the service as well
	if len(octopus.channelHandlers[service]) == 0 {
		delete(octopus.channelHandlers, service)
		octopus.stanClients[service].Close()
		delete(octopus.stanClients, service)
	}

	return nil
}

// Configure initiates a STAN context
func (octopus *EventOctopus) Configure() error {
	var (
		err error
		db  *sql.DB
	)

	octopus.configOnce.Do(func() {
		//if octopus.Config.Mode == "server" {
		db, err = sql.Open("sqlite3", octopus.Config.Connectionstring)
		defer db.Close()
		if err != nil {
			return
		}

		// 1 ping
		err = db.Ping()
		if err != nil {
			return
		}

		// migrate
		err = octopus.RunMigrations(db)
		if err != nil {
			return
		}
		//}
	})

	return err
}

// RunMigrations runs all new migrations in order
func (octopus *EventOctopus) RunMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})

	// wrap assets into Resource
	s := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
		})

	d, err := bindata.WithInstance(s)

	if err != nil {
		return err
	}

	// run migrations
	m, err := migrate.NewWithInstance("go-bindata", d, "test", driver)

	if err != nil {
		return err
	}

	err = m.Up()

	if err != nil && err.Error() != "no change" {
		return err
	}

	logrus.Debugf("Migrations ran")

	return nil
}

func (octopus *EventOctopus) nats() error {
	opts := natsServer.GetDefaultOptions()
	opts.Debug = false
	opts.Trace = false
	//opts.StoreType = stores.TypeFile
	opts.StoreType = stores.TypeMemory
	opts.FilestoreDir = "./temp"
	opts.ID = "nuts"

	sopts := natsServer.DefaultNatsServerOptions
	sopts.Port = octopus.Config.NatsPort

	var err error

	octopus.stanServer, err = natsServer.RunServerWithOpts(opts, &sopts)
	octopus.stanServer.ClusterID()
	if err != nil {
		return fmt.Errorf("Unable to start Nats-streaming server: %v", err)
	}

	logrus.Infof("Stan server started at %s:%d with ID: %v", sopts.Host, sopts.Port, octopus.stanServer.ClusterID())

	return err
}

// Start starts the receiver socket in a go routine
func (octopus *EventOctopus) Start() error {
	var err error

	// gorm db connection
	if octopus.Db, err = gorm.Open("sqlite3", octopus.Config.Connectionstring); err != nil {
		return err
	}

	// logging
	octopus.Db.SetLogger(logrus.StandardLogger())

	// natsServer startup
	if err = octopus.nats(); err != nil {
		return err
	}

	// event store client
	return octopus.eventStoreClient()
	//return nil
}

func (octopus *EventOctopus) Client(clientID string) (natsClient.Conn, error) {
	return natsClient.Connect(
		"nuts",
		clientID,
		natsClient.NatsURL(fmt.Sprintf("nats://localhost:%d", octopus.Config.NatsPort)),
	)
}

func (octopus *EventOctopus) eventStoreClient() error {
	logrus.Tracef("Connecting to Stan-Streaming server @ nats://localhost:%d", octopus.Config.NatsPort)

	sc, err := octopus.Client("event-store")
	if err != nil {
		return err
	}
	// Subscribe with manual ack mode
	// todo store Subscription?
	_, err = sc.Subscribe(ChannelConsentRequest, func(msg *natsClient.Msg) {
		event := Event{}
		// Unmarshal JSON that represents the Order data
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			logrus.Errorf("Error unmarshalling event: %v", err)
			return
		}
		// Handle the message
		logrus.Debugf("Received event [%d]: %+v\n", msg.Sequence, event)

		if err := octopus.SaveOrUpdate(event); err != nil {
			logrus.Errorf("Failed to store event: %v", err)
		}
		_ = msg.Ack() // Manual ACK
	}, natsClient.DurableName("consent-request-durable"),
		natsClient.MaxInflight(1),
		natsClient.SetManualAckMode(),
	)

	logrus.Infof("Connected to Stan-Streaming server @ nats://localhost:%d", octopus.Config.NatsPort)

	return err
}

func (octopus *EventOctopus) EventPublisher(clientId string) (natsClient.Conn, error) {
	return natsClient.Connect(
		"nuts",
		clientId,
		natsClient.NatsURL(fmt.Sprintf("nats://localhost:%d", octopus.Config.NatsPort)),
	)
}

// Shutdown closes the connection to the DB and the natsServer server
func (octopus *EventOctopus) Shutdown() error {
	var err error

	if octopus.stanServer != nil {
		octopus.stanServer.Shutdown()
	}

	if octopus.Db != nil {
		_ = octopus.Db.Close()
	}

	return err
}

// List returns all current events from Db
func (octopus *EventOctopus) List() (*[]Event, error) {
	events := &[]Event{}

	err := octopus.Db.Debug().Find(events).Error

	return events, err
}

// GetEvent returns single event or not based on given uuid
func (octopus *EventOctopus) GetEvent(uuid string) (*Event, error) {
	event := &Event{}

	err := octopus.Db.Debug().Where("uuid = ?", uuid).First(&event).Error

	return event, err
}

func (octopus *EventOctopus) SaveOrUpdate(event Event) error {
	// start transaction
	tx := octopus.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		tx.Rollback()
		return err
	}

	// actual query
	target := &Event{}
	// When using real DB:
	// err := eo.Db.Set("gorm:query_option", "FOR UPDATE").Where("uuid = ?", event.Uuid).First(&target).Error
	err := octopus.Db.Debug().Where("uuid = ?", event.Uuid).First(&target).Error

	if err == nil || gorm.IsRecordNotFoundError(err) {
		octopus.Db.Debug().Save(&event)
	} else {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
