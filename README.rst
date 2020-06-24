nuts-event-octopus
##################

Event service by the Nuts foundation for listening to events from different parts within a node.

.. image:: https://circleci.com/gh/nuts-foundation/nuts-event-octopus.svg?style=svg
    :target: https://circleci.com/gh/nuts-foundation/nuts-event-octopus
    :alt: Build Status

.. image:: https://readthedocs.org/projects/nuts-event-octopus/badge/?version=latest
    :target: https://nuts-documentation.readthedocs.io/projects/nuts-event-octopus/en/latest/?badge=latest
    :alt: Documentation Status

.. image:: https://codecov.io/gh/nuts-foundation/nuts-event-octopus/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/nuts-foundation/nuts-event-octopus

.. image:: https://api.codeclimate.com/v1/badges/704d01f281e0b8f5be33/maintainability
   :target: https://codeclimate.com/github/nuts-foundation/nuts-event-octopus/maintainability
   :alt: Maintainability

The event listener is written in Go and should be part of nuts-go as an engine.

Dependencies
************

This projects is using go modules, so version > 1.12 is recommended. 1.10 would be a minimum. ZeroMQ is used for listening to the events from the nuts-consent-bridge. Follow installation from the ZMQ website: http://zeromq.org/intro:get-the-software

Running tests
*************

Tests can be run by executing

.. code-block:: shell

    go test ./...

Generating code
***************

.. code-block:: shell

    oapi-codegen -generate server,types -package api docs/_static/nuts-event-store.yaml > api/generated.go

Generating Mock
***************

When making changes to the client interface run the following command to regenerate the mock:

.. code-block:: shell

    mockgen -destination=mock/mock_client.go -package=mock -source=pkg/events.go


Building
********

This project is part of https://github.com/nuts-foundation/nuts-go. If you do however would like a binary, just use ``go build``.

The  server API is generated from the nuts-consent-store open-api spec:

.. code-block:: shell

    oapi-codegen -generate server,types -package api docs/_static/nuts-event-store.yaml > api/generated.go

Binary format migrations
------------------------

The database migrations are packaged with the binary by using the ``go-bindata`` package.

.. code-block:: shell

    NOT_IN_PROJECT $ go get -u github.com/go-bindata/go-bindata/...
    nuts-consent-store $ cd migrations && go-bindata -pkg migrations .

README
******

The readme is auto-generated from a template and uses the documentation to fill in the blanks.

.. code-block:: shell

    ./generate_readme.sh

This script uses ``rst_include`` which is installed as part of the dependencies for generating the documentation.

Documentation
*************

To generate the documentation, you'll need python3, sphinx and a bunch of other stuff. See :ref:`nuts-documentation-development-documentation`
The documentation can be build by running

.. code-block:: shell

    /docs $ make html

The resulting html will be available from ``docs/_build/html/index.html``

Configuration
*************

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

