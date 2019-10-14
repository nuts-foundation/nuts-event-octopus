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

// ConfigRetryInterval defines the string for the flagset
const ConfigRetryInterval = "retryInterval"

// ConfigNatsPort defines the string for the flagset
const ConfigNatsPort = "natsPort"

// ConfigConnectionstring defines the string for the flagset
const ConfigConnectionstring = "connectionstring"

// ConfigRetryIntervalDefault defines the default for the nats retryInterval
const ConfigRetryIntervalDefault = 60

// ConfigNatsPortDefault defines the default nats port
const ConfigNatsPortDefault = 4222

// ConfigConnectionStringDefault defines the default sqlite connection string
const ConfigConnectionStringDefault = "file::memory:?cache=shared"

// EventOctopusConfig holds the config for the EventOctopusInstance
type EventOctopusConfig struct {
	RetryInterval    int
	NatsPort         int
	Connectionstring string
}

// IEventPublisher defines the Publish signature so it can be mocked or implemented for another tech
type IEventPublisher interface {
	Publish(subject string, event Event) error
}

// EventOctopusClient is the client interface for publishing events
type EventOctopusClient interface {
	EventPublisher(clientId string) (IEventPublisher, error)
	Subscribe(service, subject string, callbacks map[string]EventHandlerCallback) error
}

// ChannelHandlers store all the handlers for a specific channel subscription
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
var oneInstance = &sync.Once{}

// EventOctopusInstance returns the EventOctopus singleton
func EventOctopusInstance() *EventOctopus {
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
		stanClient, err := octopus.Client(service)
		if err != nil {
			return err
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

// Unsubscribe from a service and subject. If no subjects for a service are left, it closes the stanClient
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
		_ = octopus.stanClients[service].Close()
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
	sopts.Host = "0.0.0.0"
	sopts.Port = octopus.Config.NatsPort

	var err error

	octopus.stanServer, err = natsServer.RunServerWithOpts(opts, &sopts)
	if err != nil {
		return fmt.Errorf("Unable to start Nats-streaming server: %v", err)
	}
	octopus.stanServer.ClusterID()

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

// Client gets an existing or creates a new natsClient
func (octopus *EventOctopus) Client(clientID string) (natsClient.Conn, error) {
	if client, ok := octopus.stanClients[clientID]; ok {
		return client, nil
	}

	client, err := natsClient.Connect(
		"nuts",
		clientID,
		natsClient.NatsURL(fmt.Sprintf("nats://localhost:%d", octopus.Config.NatsPort)),
	)
	if err == nil {
		octopus.stanClients[clientID] = client
	}
	return client, err
}

// EventPublisher is a small wrapper around a natsClient so the user can pass an Event to Publish instead of a []byte
type EventPublisher struct {
	conn natsClient.Conn
}

// Publish accepts an Event, than marshals and publishes it at the subject choice
func (p EventPublisher) Publish(subject string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.conn.Publish(subject, data)
}

// EventPublisher gets a connection and creates a new EventPublisher
func (octopus *EventOctopus) EventPublisher(clientId string) (IEventPublisher, error) {
	conn, err := octopus.Client(clientId)
	if err != nil {
		return nil, err
	}
	return &EventPublisher{conn: conn}, nil
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
			logrus.WithError(err).Error("Failed to store event", err)
		}
		_ = msg.Ack() // Manual ACK
	}, natsClient.DurableName("consent-request-durable"),
		natsClient.MaxInflight(1),
		natsClient.SetManualAckMode(),
	)

	logrus.Infof("Connected to Stan-Streaming server @ nats://localhost:%d", octopus.Config.NatsPort)

	return err
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

	// reset the sync.once so a new connection can be created
	oneInstance = new(sync.Once)

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

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return event, err
}

// GetEventByExternalId returns single event or not based on given uuid
func (octopus *EventOctopus) GetEventByExternalId(externalID string) (*Event, error) {
	event := &Event{}

	err := octopus.Db.Debug().Where("external_id = ?", externalID).First(&event).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return event, err
}

// SaveOrUpdate saves or update the event in the store.
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
