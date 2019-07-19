.. _nuts-event-octopus-configuration:

Nuts event octopus configuration
################################

.. marker-for-readme

The following configuration parameters are available for the event service.

===================================     ======================================  ========================================
Key                                     Default                                 Description
===================================     ======================================  ========================================
events.ConfigConnectionstring           file:not_used?mode=memory&cache=shared  db connection string for event store
events.natsPort                         4222                                    Port for Nats to bind on
events.retryInterval                    60                                      Retry delay in seconds for reconnecting
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