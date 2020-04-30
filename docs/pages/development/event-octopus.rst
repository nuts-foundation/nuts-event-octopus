.. _nuts-event-octopus-development:

Nuts event octopus development
##############################

.. marker-for-readme

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
