.. _nuts-event-octopus-events-spec:

Service space event specification
#################################

Below is the flow diagram of how a consent request is transformed to a :ref:`distributed-privacy-consent` record. Blocks represent actions or commands and the hexagons represent events. The event model and the different events are descirbed further down.

.. raw:: html
    :file: ../../_static/images/consent_request_flow.svg

Event model
===========

.. code-block:: yaml

    event:
        UUID: string                   # V4 UUID
        name: string                   # event name, see table
        retryCount: int                # 0 to X
        externalId: string             # ID calculated by crypto using BSN and private key of custodian
        consentId: string              # V4 UUID assigned by Corda to a record, either a ConsentBranch or ConsentState
        initiatorLegalEntity: string   # urn style identifier of the initiating legalEntity, used to select the party who's finalizing the request
        transactionId: string          # V4 UUID identifying a possible Corda transaction that was started by this event chain
        payload: string                # Base64 encoded NewConsentRequestState JSON as accepted by consent-bridge (:ref:`nuts-consent-bridge-api`)
        error: string                  # error reason in case of a functional error

Payload per event
-----------------

+------------------------------------------+---------------------+-------------------------------------------------------------------------------------------------------------+
| Event name                               | Payload             | Description                                                                                                 |
+==========================================+=====================+=============================================================================================================+
| ConsentRequest constructed               | FullConsentRequest  | As defined by consent bridge, the attachment signatures will be an empty list                               |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest in flight                 |                     |                                                                                                             |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest flow closed               |                     |                                                                                                             |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest flow errored              |                     |                                                                                                             |
+------------------------------------------+---------------------+-------------------------------------------------------------------------------------------------------------+
| Distributed ConsentRequest received      | FullConsentRequest  | In principle the same as the NewConsentRequest above, but it'll contain AttachmentSignatures when available |
+------------------------------------------+                     | This event will also be the starting point for any other node than the initiating one                       |
| All signatures present                   |                     |                                                                                                             |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest in flight for final state |                     |                                                                                                             |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest valid                     |                     |                                                                                                             |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest acked                     |                     |                                                                                                             |
+------------------------------------------+                     |                                                                                                             |
| ConsentRequest nacked                    |                     |                                                                                                             |
+------------------------------------------+---------------------+-------------------------------------------------------------------------------------------------------------+
| Attachment signed                        | AttachmentSignature | A single signature will be present in the event. When processed by Corda, the NewConsentRequest will be     |
+------------------------------------------+                     | back with this signature included                                                                           |
| ConsentRequest flow errored              |                     |                                                                                                             |
+------------------------------------------+---------------------+-------------------------------------------------------------------------------------------------------------+
| Consent distributed                      | ConsentState        | The final consent state                                                                                     |
+------------------------------------------+---------------------+-------------------------------------------------------------------------------------------------------------+


Event types
===========

ConsentRequest constructed
--------------------------

A new consent request has been POSTed to the *consent-logic* module. It is checked, an externalId is added and it is converted to a FHIR model. The request is encrypted with the pub keys of all recipients and the published as an event with state ``consentRequest constructed``.

ConsentRequest in flight
------------------------
``consentRequest constructed`` Events are handled by the *corda-bridge*. The *corda-bridge* find the nodes to be involved and submits a transaction to Corda. The transactionId from Corda is added to a new event with state ``ConsentRequest in flight`` It'll then start polling for the initiated transaction to give feedback about its state. If all nodes have signed (including the notary), the *corda-bridge* publishes the event with added consentId and state: ``Distributed ConsentRequest received``.

ConsentRequest flow closed
--------------------------
Event that is published when a Corda close flow has been executed. This is the case when another node nacks the request. The error field will give information about the closure reason.

ConsentRequest flow errored
---------------------------
Event that is published when a Corda flow was closed with an error reason. The error field will give information about why it failed.

ConsentRequest flow success
---------------------------
Event that is usefull to debug the current state of the event flow. Will be visible in the consent store. It'll not be input for further processing, since that will be initiated by Corda events.

Distributed ConsentRequest received
-----------------------------------
The *corda-bridge* receives events from Corda when transactions are completed. It'll find the corresponding event with state: ``ConsentRequest in flight`` or ``ConsentRequest flow success`` or when another node initiated the transaction, it'll create a new event from scratch. Either way a new event with state: ``Distributed ConsentRequest received`` is created.

All signature present
---------------------
``Distributed ConsentRequest received`` events are processed by the logic module. If all signatures are present, it'll generate an event with state ``All signatures present``.

ConsentRequest in flight for final state
----------------------------------------
When a consent request is nacked or when the initiator has concluded all signatures are present, the correct flows are called by the bridge and an event is published: ``ConsentRequest in flight for final state``. This indicates that no further logical processing is needed.

ConsentRequest valid
--------------------
``Distributed ConsentRequest received`` events are processed by the logic module.  If not all signatures are present, it'll validate the record and check if all current signatures belong to the involved parties. When ok, a ``ConsentRequest valid`` event is published. This event is picked up by the logic module and auto-acked (for example when this node == the initiator) or the event must be picked up by *vendor space* for manual acking.

.. note::

    can Corda do this check in the contract using an Oracle in the form of the registry? `On Github <https://github.com/nuts-foundation/nuts-consent-cordapp/blob/master/contract/src/main/kotlin/nl/nuts/consent/contract/ConsentContract.kt#L165>`_

ConsentRequest acked
--------------------
Either the logic module or from *vendor space* an ``ConsentRequest acked`` event is produced indicating that the subject is indeed a patient in care by the given legalIdentity.


Attachment signed
-----------------
``ConsentRequest acked`` events are picked up by the logic module and a signature is produced. This will result in a ``Attachment signed`` event. This event is picked up by the bridge which will initiate an AcceptConsentRequest flow. This will result in an ``ConsentRequest in flight`` event. From here-on the event flow tree is reused.

Consent distributed
-------------------
After ``ConsentRequest in flight for final state`` Corda will transform the ``ConsentRequestState`` to a ``ConsentState``. This event is picked up by the bridge to publish a ``Consent distributed`` event.

Completed
---------
From the ``Consent distributed`` event, consent records are persisted in the *consent-store*. The event chain is completed and will enter the ``completed`` state.

Error
-----

If for some reason, an event enters the error state, the error field of the event will show the explanation. Since the event log is a circular log, errored events will not survive restarts if they are older than X (depending on the log size). It is recommended to store errored events by parsing the regular error logs and storing them somewhere. An error event published to the error channel will not be propagated across nodes, an error event published to the regular channel will be picked up an synchronized across nodes.

Channels and queues
===================

Most messaging/queueing technologies share the notion of the separation of channel and queues. Message are published to channels and stored in queues.
All queues are durable which means they will survive a restart/crash.

| Channel               | Queue                  | Consumer       | Description                                                                                             |
+=======================+========================+================+=========================================================================================================+
| consentRequest        | consentRequest         | eventStore     | The event store processes all events and stores the current state in a db                               |
|                       +------------------------+----------------+---------------------------------------------------------------------------------------------------------+
|                       | consentRequest         | consentLogic   | The validation module only processes new events and checks if they are correct                          |
|                       +------------------------+----------------+---------------------------------------------------------------------------------------------------------+
|                       | consentRequest         | consentBridge  | The bridge listens to events that are ready to send to Corda                                            |
|                       +------------------------+----------------+---------------------------------------------------------------------------------------------------------+
|                       | consentRequest         | consentStore   | The consent store handles events that are finalized and can be stored in a persistent data store        |
+-----------------------+------------------------+----------------+---------------------------------------------------------------------------------------------------------+
| consentRequestRetry   | consentRequestRetry    | eventOctopus   | General retry queue where events to be retried are sorted                                               |
+-----------------------+------------------------+----------------+---------------------------------------------------------------------------------------------------------+
| consentRequestRetry-X | consentRequestRetry-X  | eventOctopus   | Where X is the retryCount. Events are picked up and the service sleeps untill the event can be          |
|                       |                        |                | re-published to the consentRequest channel                                                              |
+-----------------------+------------------------+----------------+---------------------------------------------------------------------------------------------------------+

Retry mechanism
===============

Some errors may be caused by timeouts or poorly working infrastructure. To remedy this, events can be retried. Events that should be retried must be published to the `consentRequestRetry` channel.
The *eventOctopus* will sort the event to a different queue based on the `retryCount` (or save it as an error if the max count has been reached).
Events that are picked up by the different retry queues will remain there until the timeout has been reached (by means of delay acking the message).
The different queues have a different waiting time till the events are republished to the main channel. This can be configured by the `maxRetryCount` and `incrementalBackoff` config variables.
The `incrementalBackoff` multiplies the waiting time of the previous queue.
The default settings of 5 retries and an incremental backoff of 8 means that the waiting times for the different queues are: 1s, 8s, 64s, 512s, 4096s or 1s, 8s, ~1m, ~8m, ~1:08h.

Implementation
==============

`Nats <https://nats.io/>`_ is used as messaging system with `Nats-streaming <https://nats-io.github.io/docs/nats_streaming/intro.html>`_ as event log. The event store will be implemented with an in-memory SQLite DB.
The *Nats* service is part of the *nuts-event-octopus* and is embedded within the ``nuts`` service executable.