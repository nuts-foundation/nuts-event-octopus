.. _nuts-event-octopus-configuration:

Nuts event octopus configuration
################################

.. marker-for-readme

The following configuration parameters are available for the event service.

===================================     ====================    ================================================================================
Key                                     Default                 Description
===================================     ====================    ================================================================================
events.eventStartEpoch                  0                       Epoch at which the event stream from the consent bridge should start at
events.bridgeHost                       127.0.0.1               Host where nuts-consent-bridge is located
events.queuePort                        5563                    Port for ZMQ tcp connection
events.restPort                         8080                    Port for REST service to start event stream
events.retryInterval                    60                      Retry delay in seconds for reconnecting
===================================     ====================    ================================================================================

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