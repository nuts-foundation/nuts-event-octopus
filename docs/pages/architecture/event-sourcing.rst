.. _nuts-event-octopus-event-sourcing:

Event sourcing and CQRS in service space
########################################

Creating, accepting and finalizing consent requests is a lengthy and complicated process. Besides the steps an individual node needs to take, nodes must also be aware of each other. Without this knowledge it's hard to give a user information about the state of its request.
From a technical point of view, there's also the complexity of retrying failed steps and disaster recovery. Complex problems in their own definition, but even harder when let lose in a distributed environment where acceptance of the technology depends on simplicity.

Enter `event-sourcing <https://martinfowler.com/eaaDev/EventSourcing.html>`_, a technique that stores events in a log, ready to be re-run at a later stage.
The example most commonly used is that of a purchase or order, where the order goes through different states before being completed.
Consent rules can be seen as *orders*. Together with event-sourcing, `CQRS <https://martinfowler.com/bliki/CQRS.html>`_ is commonly used in these types of systems.
The main idea is to make the entire system event-based, where different services react on different events. Events are handles by multiple services.
One of these services is responsible for creating the current state. This state can then be queried by vendor-space for the current state of a consent request.
Other services act on the events to help transition the state. In the case of :ref:`disaster-recovery` the entire event log can be replayed.

Error handling & retry
======================

A single fully accepted consent request has a 100 different ways of failing on its route to acceptance. Many of which can be a result of temporary problems such as communication problems and/or unknown technical problems. Some might occur through incorrect configuration and some might occur due to version conflicts.
Almost all of these problems can be fixed by retrying at another time. Only functional rejected requests have to be dealt with by a user.

Communication problems occur because of network hiccups or external services that are down. Hiccups can usually be fixed by retrying processing immediately. External services that suffer downtime are usually fixed within a minute when a hot-standby HA method is chosen or within 15 minutes if an administrator has to deal with it. In badly maintained environments this can be stretched to several hours.

Problems that occur due to misconfiguration, like an incorrect address should be spotted right after an update or after installation and are expected to be fixed within an hour.

Version conflicts can occur when another node is running a newer version than the node giving the error. In Nuts, this is avoided by updating the internal processing and data model in a version before any interfaces are changed. This means that nodes can handle newer version before they are actually used. Two major versions are allowed to be active at any given time in the network (allowing for a transition time). If for some reason a node falls two versions behind and is thus unable to process newer events, a retry in a week or so would fix this, since this will probably only occur in test networks. A test network will be updated quicker than a production network.

The retry mechanism is implemented as a single channel of events per retry interval. So 1 channel for 1 second retries, 1 for 1 minute and so on. The increase in interval and the maximum number of retries is configurable. The recommendation would be to do something like: 1s, 1m, 15m, 1h, 24h, 7d.
The retry channels/queues are not durable, meaning they won't survive a restart of the node. When restarted the :ref:`disaster-recovery` mechanism will retry any non-completed events anyway.

An event that results in an error condition is *acked* and then published to the correct retry channel. Failing to publish to the retry channel (or to the event log in general) is a **Fatal** error.

Encryption
==========

Because the events hold the actual consent data (and thus personal information), it can only be stored to the event log **after** it has been encrypted.
All events are processed in-memory except the event log. Encrypting the data before the first event makes sure that records written to disk are encrypted.
This means that the first REST call for creating a new consent record will be synchronous until the encryption step.
After encryption it'll be an event and the event UUID can be returned to the caller.

.. _disaster-recovery:

Disaster recovery
=================

In the case all Nuts nodes have failed and the entire system has to be started again, the event log can be used to resume consent requests.
This is done by first starting the event store (nuts-event-octopus module) which will re-read the entire event log.
This will recreate an internal db (SQLite) with the current state of each consent request. Any vendor-space services can now already happily query request states.
After the event store has been recreated, the other modules will be started that are depended on events. Each event is processed by each module.
Modules only react to their events they are interested in and they will check with the consent store if the given state is still current, if not the event is skipped.

.. note::

    An event log is a ring-based log and therefore has a maximum size. This size has to be chosen wisely, for normal operations it can be expected that at most 1 consent request per second is handled. Each request has to go through 6 states or so. So for storing a day of logs, a log size of `6 * 24 * 60 * 60 = 518400`. This will account for downtime of the entire nuts network for a full day. This is highly unlikely, these log sizes do, however, are needed when doing bulk imports. For example, a hospital with 100.000 active consent records, will need a log size of 600.000 to do a full bulk import. (probably less, but better safe than sorry).

.. note::

    In the future the last processed event ID can be persisted, so modules can skip the first X events.
