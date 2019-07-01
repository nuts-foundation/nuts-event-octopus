nuts-event-octopus
##################

Event service by the Nuts foundation for listening to events from different parts within a node.

.. image:: https://travis-ci.org/nuts-foundation/nuts-event-octopus.svg?branch=master
    :target: https://travis-ci.org/nuts-foundation/nuts-event-octopus
    :alt: Build Status

.. image:: https://readthedocs.org/projects/nuts-event-octopus/badge/?version=latest
    :target: https://nuts-documentation.readthedocs.io/projects/nuts-event-octopus/en/latest/?badge=latest
    :alt: Documentation Status

.. image:: https://codecov.io/gh/nuts-foundation/nuts-event-octopus/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/nuts-foundation/nuts-event-octopus

The event listener is written in Go and should be part of nuts-go as an engine.

Dependencies
************

This projects is using go modules, so version > 1.12 is recommended. 1.10 would be a minimum. ZeroMQ is used for listening to the events from the nuts-consent-bridge. Follow installation from the ZMQ website: http://zeromq.org/intro:get-the-software

Running tests
*************

Tests can be run by executing

.. code-block:: shell

    go test ./...

Building
********

This project is part of https://github.com/nuts-foundation/nuts-go. If you do however would like a binary, just use ``go build``.


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

