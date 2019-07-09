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

package engine

import (
	"github.com/nuts-foundation/nuts-event-octopus/pkg"
	engine "github.com/nuts-foundation/nuts-go/pkg"
	"github.com/spf13/pflag"
)

// NewEventOctopusEngine creates the engine configuration for nuts-go.
func NewEventOctopusEngine() *engine.Engine {
	i := pkg.EventOctopusIntance()

	return &engine.Engine{
		Config:    &i.Config,
		ConfigKey: "events",
		Configure: i.Configure,
		FlagSet:   flagSet(),
		Name:      "Events octopus",
		Start:	   i.Start,
		Shutdown:  i.Shutdown,
	}
}

func flagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("event octopus", pflag.ContinueOnError)

	flags.Int(pkg.ConfigEpoch, pkg.ConfigEpochDefault, "Epoch at which the event stream from the consent bridge should start at")
	flags.String(pkg.ConfigZmqAddress, pkg.ConfigZmqAddressDefault, "ZeroMQ address of the consent-bridge")
	flags.Int(pkg.ConfigRetryInterval, pkg.ConfigRetryIntervalDefault, "Retry delay in seconds for reconnecting")

	return flags
}