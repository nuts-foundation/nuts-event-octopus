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
	ConsentId            string `json:"consentId"`
	TransactionId        string `json:"transactionId"`
	InitiatorLegalEntity string `gorm:"not null";json:"initiatorLegalEntity"`
	Error                *string `json:"error"`
	ExternalId           string `gorm:"not null";json:"externalId"`
	Payload              string `gorm:"not null";json:"payload"`
	RetryCount           int32  `json:"retryCount"`
	Name                 string `gorm:"not null";json:"name"`
	Uuid                 string `gorm:"PRIMARY_KEY";json:"uuid"`
}

func (e Event) String() string {
	return fmt.Sprintf("Name: %v, uuid: %v, externalId: %v, retryCount: %v, error: %v", e.Name, e.Uuid, e.ExternalId, e.RetryCount, e.Error )
}

type EventHandlerCallback func(event *Event)

const EventConsentRequestConstructed = "consentRequest constructed"
const EventConsentRequestInFlight = "consentRequest in flight"
const EventConsentRequestFlowErrored = "consentRequest flow errored"
const EventConsentRequestFlowSuccess = "consentRequest flow success"
const EventDistributedConsentRequestReceived = "distributed ConsentRequest received"
const EventAllSignaturesPresent = "all signatures present"
const EventInFinalFlight = "consentRequest in flight for final state"
const EventConsentRequestValid = "consentRequest valid"
const EventConsentRequestAcked = "consentRequest acked"
const EventConsentRequestNacked = "consentRequest nacked"
const EventAttachmentSigned = "attachment signed"
const EventConsentDistributed = "consent distributed"
const EventCompleted = "completed"
const EventErrored = "error"
const ChannelConsentRequest = "consentRequest"
const ChannelConsentRetry = "consentRequestRetry"
const ChannelConsentErrored = "consentRequestErrored"
