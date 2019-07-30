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

// Event is the type used for Gorm
type Event struct {
	ConsentId            string
	TransactionId        string
	InitiatorLegalEntity string `gorm:"not null"`
	Error                *string
	ExternalId           string `gorm:"not null"`
	Payload              string `gorm:"not null"`
	RetryCount           int32
	Name                 string `gorm:"not null"`
	Uuid                 string `gorm:"PRIMARY_KEY"`
}

type EventHandlerCallback func(event *Event)

const ChannelConsentRequest = "consentRequest"
const ChannelConsentRetry = "consentRequestRetry"
const ChannelConsentErrored = "consentRequestErrored"

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
