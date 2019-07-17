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

package api

import (
	"github.com/nuts-foundation/nuts-event-octopus/pkg"
)

func Convert(e pkg.Event) Event {
	return Event{
		Error:      e.Error,
		Custodian:  Identifier(e.Custodian),
		ConsentId:  e.ConsentId,
		ExternalId: e.ExternalId,
		State:      e.State,
		Payload:    e.Payload,
		RetryCount: e.RetryCount,
		Uuid:       e.Uuid,
	}
}

func ConvertList(e *[]pkg.Event) []Event {
	events := make([]Event, len(*e))

	for i, el := range *e {
		events[i] = Convert(el)
	}

	return events
}
