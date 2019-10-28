.. _nuts-event-octopus-configuration:

Nuts event octopus configuration
################################

.. marker-for-readme

The following configuration parameters are available for the event service.

===================================     ======================================  ========================================
Key                                     Default                                 Description
===================================     ======================================  ========================================
events.ConfigConnectionstring           file::memory:?cache=shared              db connection string for event store
events.natsPort                         4222                                    Port for Nats to bind on
events.retryInterval                    60                                      Retry delay in seconds for reconnecting
events.autoRecover                      true                                    Republish unfinished events at startup
events.purgeCompleted                   true                                    Purge completed events at startup
events.maxRetryCount                    5                                       Max number of retries for events before giving up (only for recoverable errors)
events.incrementalBackoff               8                                       Incremental backoff per retry queue, queue 0 retries after 1 second, queue 1 after {incrementalBackoff} * {previousDelay}
===================================     ======================================  ========================================

As with all other properties for nuts-go, they can be set through yaml:

.. sourcecode:: yaml

    events:
       eventStartEpoch: 0

as commandline property

.. sourcecode:: shell

    ./nuts --events.eventStartEpoch 0

Or by using environment variables

.. sourcecode:: shell

    NUTS_EVENTS_EVENTSTARTEPOCH=0 ./nuts