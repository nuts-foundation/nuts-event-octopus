.. _nuts-event-octopus-events-spec:

Service space event specification
#################################

Event model
===========

.. code-block:: yaml

    event:
        UUID: string        # V4 UUID
        state: string       # ['requested', 'offered', 'to be accepted', 'accepted', 'finalized', 'to be persisted', 'completed', 'error']
        retryCount: int     # 0 to X
        externalId: string  # ID calculated by crypto using BSN and private key of custodian
        consentId: string   # V4 UUID assigned by Corda to a record
        custodian: string   # urn style identifier of the custodian, used to select correct key for decryption
        payload: string     # Base64 encoded NewConsentRequestState JSON as accepted by consent-bridge (:ref:`nuts-consent-bridge-api`)
        error: string       # error reason in case of a functional error


States
======

Requested
---------

A new consent request has been POSTed to the *consent-logic* module. It is checked, an externalId is added and it is converted to a FHIR model. The request is encrypted with the pub keys of all recipients and the published as an event with state ``requested``.

Offered
-------
``requested`` Events are handles by the *corda-bridge*. The *corda-bridge* find the nodes to be involved and submits a transaction to Corda. If all nodes have signed (including the notary), the *corda-bridge* publishes the event with added consentId and state: ``offered``.

To be accepted
--------------
the *corda-bridge* received events from Corda when new transactions are completed. It'll find the corresponding event with state: ``offered`` or when another node initiated the transaction, it'll create a new event from scratch. Either way a new event with state: ``to be accepted`` is created. Corresponding events with state ``accepted`` are re-published with the same state (to be finalized).

Accepted
--------
``to be accepted`` events are picked up by the *consent-logic* module and are checked contextually and syntactically. If anything is wrong it'll ``error`` the event, signalling the bridge to cancel the consent request in Corda. If all is well, the event can be accepted (with the ``auto-ack`` configuration) or it can be picked up by an UI in vendor-space to accept at a later stage. When accepted the event will enter the ``accepted`` state.

Finalized
---------
Consent requests that haven been accepted by all recipients will be checked by the *consent-logic* module to see if the public-keys in the signatures match with the given *legalIdentities* using the *registry*. If all is well it'll signal the bridge to finalize the transaction. Finalizing is a synchronous action on the *consent-bridge*. When done the event enters the ``finalized`` state.

.. note::

    can Corda do this check in the contract using an Oracle in the form of the registry? `On Github <https://github.com/nuts-foundation/nuts-consent-cordapp/blob/master/contract/src/main/kotlin/nl/nuts/consent/contract/ConsentContract.kt#L165>`_

To be persisted
---------------
Consent requests that are finalized are consumed in Corda and a new Consent state is produced. The *consent-bridge* will publish an event to persist the new consent rule in the *consent-store*. In case of a disaster-recovery scenarion, the *consent-bridge* must also process the ``finalized`` events and publish a ``to be persisted`` event if Corda no longer has the consent request record but does contain the consent record.

Completed
---------
When consent records are persisted in the *consent-store* the event is completed and will enter the ``completed`` state.

Error
-----

If for some reason, an event enters the error state, the error field of the event will show the explanation. Since the event log is a circular log, errored events will not survive restarts if they are older than X (depending on the log size). It is recommended to store errored events by parsing the regular error logs and storing them somewhere.

Channels and queues
===================

Most messaging/queueing technologies share the notion of the separation of channel and queues. Message are published to channels and stored in queues.

+----------------+------------+----------------+---------------------------------------------------------------------------------------------------------+
| Channel        | Queue      | Consumer       | Description                                                                                             |
+================+============+================+=========================================================================================================+
| consentRequest | eventStore | eventStore     | The event store processes all events and stores the current state in a db                               |
|                +------------+----------------+---------------------------------------------------------------------------------------------------------+
|                | validation | consentLogic   | The validation module only processes new events and checks if they are correct                          |
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