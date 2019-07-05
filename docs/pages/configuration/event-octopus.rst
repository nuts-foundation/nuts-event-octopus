.. _nuts-event-octopus-configuration:

Nuts event octopus configuration
################################

.. marker-for-readme

The following configuration parameters are available for the event service.

===================================     =====================    ================================================================================
Key                                     Default                  Description
===================================     =====================    ================================================================================
events.eventStartEpoch                  0                        Epoch at which the event stream from the consent bridge should start at
events.zmqAddress                       tcp://127.0.0.1:5563     ZeroMQ address of the consent-bridge
events.restAddress                      http://localhost:8080    REST address of consent-bridge
events.retryInterval                    60                       Retry delay in seconds for reconnecting
===================================     =====================    ================================================================================

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