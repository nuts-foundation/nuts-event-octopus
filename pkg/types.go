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
	ConsentId  string `gorm:"not null"`
	Custodian  string `gorm:"not null"`
	Error      *string
	ExternalId string `gorm:"not null"`
	Payload    string `gorm:"not null"`
	RetryCount int32
	State      string `gorm:"not null"`
	Uuid       string `gorm:"PRIMARY_KEY"`
}

const EventStateRequested = "requested"
const EventStateOffered = "offered"
const EventStateToBeAccepted = "to be accepted"
const EventStateAccepted = "accepted"
const EventStateFinalized = "finalized"
const EventStateToBePersisted = "to be persisted"
const EventStateCompleted = "completed"
const EventStateToBeError = "error"
