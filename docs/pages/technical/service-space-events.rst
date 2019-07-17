.. _nuts-event-octopus-events-spec:

Service space event specification
#################################

Event model
===========

.. code-block:: yaml

    event:
        UUID: string        # V4 UUID
        state: string       # ['requested', 'offered', 'to be accepted', 'accepted', 'finalized', 'complete']
        retryCount: int     # 0 to X
        externalId: string  # ID calculated by crypto using BSN and private key of custodian
        consentId: string   # V4 UUID assigned by Corda to a record
        custodian: string   # urn style identifier of the custodian, used to select correct key for decryption
        payload: string     # NewConsentRequestState JSON as accepted by consent-bridge (:ref:`nuts-consent-bridge-api`)



Channels and queues
===================

Most messaging/queueing technologies share the notion of the separation of channel and queues. Message are published to channels and stored in queues.

+----------------+------------+----------------+---------------------------------------------------------------------------------------------------------+
| Channel        | Queue      | Consumer       | Description                                                                                             |
+================+============+================+=========================================================================================================+
| consentRequest | eventStore | eventStore     | The event store processes all events and stores the current state in a db                               |
|                +------------+----------------+---------------------------------------------------------------------------------------------------------+
|                | validation | fhirValidation | The validation module only processes new events and checks if they are correct                          |
|                +------------+----------------+---------------------------------------------------------------------------------------------------------+
|                | bridge     | consentBridge  | The bridge listens to events that are ready to send to Corda                                            |
|                +------------+----------------+---------------------------------------------------------------------------------------------------------+
|                | store      | consentStore   | The consent store handles events that are finalized and can be stored in a persistent data store        |
+----------------+------------+----------------+---------------------------------------------------------------------------------------------------------+
| retryX         | retryX     | eventOctopus   | Where X is the retryCount. Events are picked up and the service sleeps untill the event can be          |
|                |            |                | re-published                                                                                            |
+----------------+------------+----------------+---------------------------------------------------------------------------------------------------------+

Implementation
==============

`Nats <https://nats.io/>`_ is used as messaging system with `Nats-streaming <https://nats-io.github.io/docs/nats_streaming/intro.html>`_ as event log. The event store will be implemented with an in-memory SQLite DB.
The *Nats* service is part of the *nuts-event-octopus* and is embedded within the ``nuts`` service executable.