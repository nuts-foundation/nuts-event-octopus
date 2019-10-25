/*
 * Nuts event octopus
 * Copyright (C) 2019. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package pkg

import "fmt"

// Event is the type used for Gorm
type Event struct {
	ConsentID            string  `json:"consentId"`
	TransactionID        string  `json:"transactionId"`
	InitiatorLegalEntity string  `gorm:"not null" json:"initiatorLegalEntity"`
	Error                *string `json:"error"`
	ExternalID           string  `gorm:"not null" json:"externalId"`
	Payload              string  `gorm:"not null" json:"payload"`
	RetryCount           int32   `json:"retryCount"`
	Name                 string  `gorm:"not null" json:"name"`
	UUID                 string  `gorm:"PRIMARY_KEY" json:"uuid"`
}

func (e Event) String() string {
	return fmt.Sprintf("Name: %v, uuid: %v, externalId: %v, retryCount: %v, error: %v", e.Name, e.UUID, e.ExternalID, e.RetryCount, e.Error)
}

// TODO create matrix of overrides!!!

// EventHandlerCallback defines the signature of an event handler method.
type EventHandlerCallback func(event *Event)

// EventConsentRequestConstructed is the event emitted directly after consent request creation to start the flow
const EventConsentRequestConstructed = "consentRequest constructed"

// EventConsentRequestInFlight is used to indicate the node is waiting for corda to come back with more information
const EventConsentRequestInFlight = "consentRequest in flight"

// EventConsentRequestFlowErrored indicates something went wrong 😔
const EventConsentRequestFlowErrored = "consentRequest flow errored"

// EventConsentRequestFlowSuccess is used when the Corda flow has been executed successfully
const EventConsentRequestFlowSuccess = "consentRequest flow success"

// EventDistributedConsentRequestReceived is broadcasted by the consent-bridge when a request has been received
const EventDistributedConsentRequestReceived = "distributed ConsentRequest received"

// EventAllSignaturesPresent is emitted when all nodes have signed the consentRecord
const EventAllSignaturesPresent = "all signatures present"

// EventInFinalFlight indicates the consentRequest is in flight for final storage
const EventInFinalFlight = "consentRequest in flight for final state"

// EventConsentRequestValid indicates the consentRequest is technically a valid request
const EventConsentRequestValid = "consentRequest valid"

// EventConsentRequestAcked indicates the consentRequest has approved by the vendor system
const EventConsentRequestAcked = "consentRequest acked"

// EventConsentRequestNacked indicates the consentRequest has been denied by the vendor system
const EventConsentRequestNacked = "consentRequest nacked"

// EventAttachmentSigned indicates one of the signatures has been signed
const EventAttachmentSigned = "attachment signed"

// EventConsentDistributed indicates the consent request has been distributed among all nodes
const EventConsentDistributed = "consent distributed"

// EventCompleted indicates the flow has been completed 🎉
const EventCompleted = "completed"

// EventErrored indicates the flow has errored
const EventErrored = "error"

// ChannelConsentRequest is the default channel to broadcast events
const ChannelConsentRequest = "consentRequest"

// ChannelConsentRetry can be used to broadcast events which are errored and should be retried
const ChannelConsentRetry = "consentRequestRetry"

// ChannelConsentErrored can be used to broadcast events which are errored and can not be retried
const ChannelConsentErrored = "consentRequestErrored"
