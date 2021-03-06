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

.. include:: docs/pages/development/event-octopus.rst
    :start-after: .. marker-for-readme

Configuration
*************

The following configuration parameters are available:

.. include:: README_options.rst

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