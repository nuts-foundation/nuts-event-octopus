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
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-event-octopus/pkg"
)

type ApiWrapper struct {
	Eo *pkg.EventOctopus
}

func (w ApiWrapper) List(ctx echo.Context) error {
	events, err := w.Eo.List()

	if err != nil {
		return fmt.Errorf("Error during fetching list of events from DB: %v", err)
	}

	resp := EventListResponse{
		Events: ConvertList(events),
	}

	return ctx.JSON(200, resp)
}

func (w ApiWrapper) GetEvent(ctx echo.Context, uuid string) error {
	event, err := w.Eo.GetEvent(uuid)

	if err != nil {
		return fmt.Errorf("Error while fetching event from DB: %v", err)
	}

	if event == nil {
		return ctx.NoContent(404)
	}

	resp := Convert(*event)

	return ctx.JSON(200, resp)
}

func (w ApiWrapper) GetEventByExternalId(ctx echo.Context, externalId string) error {
	event, err := w.Eo.GetEventByExternalId(externalId)

	if err != nil {
		return fmt.Errorf("Error while fetching event from DB: %v", err)
	}

	if event == nil {
		return ctx.NoContent(404)
	}

	resp := Convert(*event)

	return ctx.JSON(200, resp)
}
