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
	"strings"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jinzhu/gorm"
	natsServer "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats-streaming-server/stores"
	natsClient "github.com/nats-io/stan.go"
	"github.com/nuts-foundation/nuts-event-octopus/migrations"
	core "github.com/nuts-foundation/nuts-go-core"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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

// ConfigAutoRecover is the config name for republishing unfinished events at startup
const ConfigAutoRecover = "autoRecover"

// ConfigPurgeCompleted is the config name for enabling purging completed events
const ConfigPurgeCompleted = "purgeCompleted"

// ConfigMaxRetryCount is the config name for the number of retries for an event
const ConfigMaxRetryCount = "maxRetryCount"

// ConfigMaxRetryCountDefault is the default setting for the number of retries
const ConfigMaxRetryCountDefault = 5

// ConfigIncrementalBackoff is the name of the config used to determine the incremental backoff
const ConfigIncrementalBackoff = "incrementalBackoff"

// ConfigIncrementalBackoffDefault is the default setting for the incremental backoff of retrying events
const ConfigIncrementalBackoffDefault = 8

// Name is the name of this module
const Name = "Events octopus"

// ClientID is the Nats client ID
const ClientID = "event-store"

// EventOctopusConfig holds the config for the EventOctopusInstance
type EventOctopusConfig struct {
	RetryInterval      int
	NatsPort           int
	Connectionstring   string
	AutoRecover        bool
	PurgeCompleted     bool
	MaxRetryCount      int
	IncrementalBackoff int
}

// GetMode derives the mode (from the global mode) the engine should run in
func (c EventOctopusConfig) GetMode() string {
	// Since this module does not support mode selection (client/server), it will just derive it from the global mode
	return core.NutsConfig().GetEngineMode("")
}

// IEventPublisher defines the Publish signature so it can be mocked or implemented for another tech
type IEventPublisher interface {
	Publish(subject string, event Event) error
}

// EventOctopusClient is the client interface for publishing events
type EventOctopusClient interface {
	EventPublisher(clientID string) (IEventPublisher, error)
	Subscribe(service, subject string, callbacks map[string]EventHandlerCallback) error
	Diagnostics() []core.DiagnosticResult
}

// ChannelHandlers store all the handlers for a specific channel subscription
type ChannelHandlers struct {
	subscription natsClient.Subscription
	handlers     map[string]EventHandlerCallback
}

// EventOctopus is the default implementation for EventOctopusInstance
type EventOctopus struct {
	Name       string
	Config     EventOctopusConfig
	configOnce sync.Once
	stanServer *natsServer.StanServer
	Db         *gorm.DB
	sqlDb      *sql.DB
	// Clients per service
	stanClients     map[string]natsClient.Conn
	channelHandlers map[string]map[string]ChannelHandlers
	// Retry
	delayedConsumers []*DelayedConsumer
}

var instance *EventOctopus
var oneInstance = &sync.Once{}
var mutex = sync.Mutex{}

// EventOctopusInstance returns the EventOctopus singleton
func EventOctopusInstance() *EventOctopus {
	oneInstance.Do(func() {
		instance = &EventOctopus{
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
		stanClient, err := octopus.client(service)
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

type dbDiagnosticResult struct {
	pingError error
}

type natsDiagnosticsResult struct {
	up        bool
	natsMode  string
	natsPort  int
	stanID    string
	lastError error
}

// Name returns the name of the natsDiagnosticsResult
func (ndr natsDiagnosticsResult) Name() string {
	return "Nats streaming server"
}

// String returns the outcome of the natsDiagnosticsResult
func (ndr natsDiagnosticsResult) String() string {
	if !ndr.up {
		return "DOWN"
	}

	lastError := "NONE"
	if ndr.lastError != nil {
		lastError = ndr.lastError.Error()
	}

	return fmt.Sprintf("mode: %s @ 0.0.0.0:%d, ID: %s, last error: %s", ndr.natsMode, ndr.natsPort, ndr.stanID, lastError)
}

// Name returns the name of the GenericDiagnosticResult
func (ddr dbDiagnosticResult) Name() string {
	return "DB"
}

// String returns the outcome of the GenericDiagnosticResult
func (ddr dbDiagnosticResult) String() string {
	if ddr.pingError == nil {
		return "ping: true"
	}

	return fmt.Sprintf("ping: false, error: %v", ddr.pingError)
}

// Diagnostics returns diagnostic reports from the nats streaming service and the DB
func (octopus *EventOctopus) Diagnostics() []core.DiagnosticResult {
	var (
		stanState core.DiagnosticResult
		dbState   core.DiagnosticResult
	)

	if octopus.stanServer == nil {
		stanState = natsDiagnosticsResult{
			up: false,
		}
	} else {
		stanState = natsDiagnosticsResult{
			up:        true,
			natsMode:  octopus.stanServer.State().String(),
			natsPort:  octopus.Config.NatsPort,
			stanID:    octopus.stanServer.ClusterID(),
			lastError: octopus.stanServer.LastError(),
		}
	}

	dbState = dbDiagnosticResult{
		pingError: octopus.sqlDb.Ping(),
	}

	return []core.DiagnosticResult{
		stanState,
		dbState,
	}
}

// Configure initiates a STAN context
func (octopus *EventOctopus) Configure() error {
	var (
		err error
	)

	octopus.configOnce.Do(func() {
		err = octopus.configure()
	})

	return err
}

func (octopus *EventOctopus) configure() error {
	var (
		err error
	)

	if octopus.Config.GetMode() != core.ServerEngineMode {
		return nil
	}

	octopus.sqlDb, err = sql.Open("sqlite3", octopus.Config.Connectionstring)

	if err != nil {
		return err
	}

	// 1 ping
	err = octopus.sqlDb.Ping()
	if err != nil {
		return err
	}

	// migrate
	err = octopus.RunMigrations(octopus.sqlDb)
	if err != nil {
		return err
	}

	return nil
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

func (octopus *EventOctopus) startStanServer() error {
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
		return fmt.Errorf("Unable to start Nats-streaming server: %w", err)
	}
	octopus.stanServer.ClusterID()

	logrus.Infof("Stan server started at %s:%d with ID: %v", sopts.Host, sopts.Port, octopus.stanServer.ClusterID())

	return err
}

// Start starts the receiver socket in a go routine
func (octopus *EventOctopus) Start() error {
	var err error

	if octopus.Config.GetMode() != core.ServerEngineMode {
		return nil
	}

	// gorm db connection
	if octopus.Db, err = gorm.Open("sqlite3", octopus.sqlDb); err != nil {
		return err
	}

	// logging
	octopus.Db.SetLogger(logrus.StandardLogger())

	// natsServer startup
	if err = octopus.startStanServer(); err != nil {
		return err
	}

	// event store client
	if err = octopus.startSubscribers(); err != nil {
		return err
	}

	if octopus.Config.AutoRecover {
		if err := octopus.recover(); err != nil {
			return err
		}
	}

	if octopus.Config.PurgeCompleted {
		if err := octopus.purgeCompleted(); err != nil {
			return err
		}
	}

	return nil
}

// client gets an existing or creates a new natsClient
func (octopus *EventOctopus) client(clientID string) (natsClient.Conn, error) {
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
func (octopus *EventOctopus) EventPublisher(clientID string) (IEventPublisher, error) {
	conn, err := octopus.client(clientID)
	if err != nil {
		return nil, err
	}
	return &EventPublisher{conn: conn}, nil
}

func (octopus *EventOctopus) startSubscribers() error {
	logrus.Tracef("Connecting to Stan-Streaming server @ nats://localhost:%d", octopus.Config.NatsPort)

	sc, err := octopus.client(ClientID)
	if err != nil {
		return err
	}
	// Subscribe to main subject
	_, err = sc.Subscribe(ChannelConsentRequest, func(msg *natsClient.Msg) {
		event := octopus.saveMsgAsEvent(msg.Data)

		// Handle the message
		logrus.Debugf("received event [%d]: %+v\n", msg.Sequence, event)
	}, natsClient.DurableName("consent-request-durable"),
		natsClient.StartWithLastReceived(),
	)
	if err != nil {
		return err
	}

	// Subscribe to error subject
	_, err = sc.Subscribe(ChannelConsentErrored, func(msg *natsClient.Msg) {
		event := octopus.saveMsgAsEvent(msg.Data)

		// Handle the message
		logrus.Debugf("received error event [%d]: %+v\n", msg.Sequence, event)
	}, natsClient.DurableName("consent-request-error-durable"),
		natsClient.StartWithLastReceived(),
	)
	if err != nil {
		return err
	}

	// Subscribe to retry subject
	_, err = sc.Subscribe(ChannelConsentRetry, func(msg *natsClient.Msg) {
		event := Event{}

		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			logrus.WithError(err).Errorf("Error unmarshalling event")
			octopus.saveMsgAsErrored(msg.Data, err.Error())

			return
		}

		if event.RetryCount >= octopus.Config.MaxRetryCount {
			event.Name = EventErrored
			errStr := "max retry count reached"
			event.Error = &errStr
			octopus.SaveOrUpdateEvent(event)

			return
		}

		if err := octopus.publishEventToRetryChannel(event); err != nil {
			logrus.WithError(err).Fatal("failed to publish message to retry channel")
		}

		logrus.Debugf("received retry event [%d]: %+v\n", msg.Sequence, event)
	}, natsClient.DurableName("consent-request-retry-durable"),
		natsClient.StartWithLastReceived(),
	)
	if err != nil {
		return err
	}

	// subscribe retry channels
	octopus.delayedConsumers = NewDelayedConsumerSet(ChannelConsentRetry, ChannelConsentRequest, octopus.Config.MaxRetryCount, time.Second, octopus.Config.IncrementalBackoff, sc)
	for _, dc := range octopus.delayedConsumers {
		if err := dc.Start(); err != nil {
			return err
		}
	}

	logrus.Infof("Connected to Stan-Streaming server @ nats://localhost:%d", octopus.Config.NatsPort)

	return err
}

func (octopus *EventOctopus) publishEventToRetryChannel(event Event) error {
	conn, err := octopus.client(ClientID)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("%s-%d", ChannelConsentRetry, event.RetryCount)
	event.RetryCount++

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// publish async otherwise we'll be waiting for the retry procedure to ack
	_, err = conn.PublishAsync(channel, eventBytes, func(s string, e error) {
		if e != nil {
			logrus.WithError(err).Error("did not recieve ack for message published to retry queue")
		}
	})

	return err
}

func (octopus *EventOctopus) publishEventToChannel(event Event, channel string) error {
	conn, err := octopus.client(ClientID)
	if err != nil {
		return err
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// publish async otherwise we'll be waiting for the retry procedure to ack
	return conn.Publish(channel, eventBytes)
}

func (octopus *EventOctopus) saveMsgAsEvent(data []byte) Event {
	event := Event{}

	err := json.Unmarshal(data, &event)
	if err != nil {
		logrus.WithError(err).Errorf("Error unmarshalling event")
		return octopus.saveMsgAsErrored(data, err.Error())
	}

	if err := octopus.SaveOrUpdateEvent(event); err != nil {
		logrus.WithError(err).Fatal("could not store event")
	}

	return event
}

func (octopus *EventOctopus) saveMsgAsErrored(bytes []byte, msg string) Event {
	event := Event{
		InitiatorLegalEntity: "unknown",
		Error:                &msg,
		ExternalID:           "unknown",
		Payload:              string(bytes),
		RetryCount:           0,
		Name:                 EventErrored,
		UUID:                 uuid.NewV4().String(),
	}

	// go through transaction
	if err := octopus.SaveOrUpdateEvent(event); err != nil {
		logrus.WithError(err).Fatal("could not store errored event")
	}

	return event
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

	if err != nil {
		return nil, err
	}

	return event, err
}

// GetEventByExternalID returns single event or not based on given uuid
func (octopus *EventOctopus) GetEventByExternalID(externalID string) (*Event, error) {
	event := &Event{}

	err := octopus.Db.Debug().Where("external_id = ?", externalID).First(&event).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return event, err
}

// SaveOrUpdateEvent saves or update the event in the store.
func (octopus *EventOctopus) SaveOrUpdateEvent(event Event) error {

	// sqlite is giving problems
	mutex.Lock()
	defer mutex.Unlock()

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
	// err := eo.Db.Set("gorm:query_option", "FOR UPDATE").Where("uuid = ?", event.UUID).First(&target).Error
	err := octopus.Db.Debug().Where("uuid = ?", event.UUID).First(&target).Error

	// TODO, check if event has to be overwritten!!!!
	if err == nil || gorm.IsRecordNotFoundError(err) {
		err = octopus.Db.Debug().Save(&event).Error
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// recover creates a map from event.UUID to event.Name
// for all items in the map that do not have the event.Name == EventCompleted, a new event will be published
// unless the max retry count has been reached.
func (octopus *EventOctopus) recover() error {
	events, err := octopus.allEvents()

	// filter out all non-completed events, in-place
	var added int
	for i, e := range events {
		if e.Name != EventCompleted {
			events[added] = events[i]
			added++
		}
	}

	events = events[0:added]

	// Nats client ID's can not contain whitespace
	publisher, err := octopus.EventPublisher(strings.Replace(octopus.Name, " ", "_", -1))
	if err != nil {
		return err
	}

	// re publish
	for _, e := range events {
		err := publisher.Publish(ChannelConsentRequest, e)
		if err != nil {
			logrus.Error("error during publishing of recovery events, stopping")
			return err
		}
	}

	if len(events) > 0 {
		logrus.Infof("Re-published %d events", len(events))
	}

	return nil
}

func (octopus *EventOctopus) allEvents() (events []Event, err error) {
	err = octopus.Db.Debug().Find(&events).Error

	return events, err
}

// purgeCompleted removes all events from the DB with name == Completed
func (octopus *EventOctopus) purgeCompleted() error {
	if err := octopus.Db.Debug().Delete(Event{}, "name = ?", EventCompleted).Error; err != nil {
		return err
	}

	logrus.Infof("Event DB purged from completed events")

	return nil
}
