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
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/nuts-foundation/nuts-event-octopus/api"
	"github.com/nuts-foundation/nuts-event-octopus/pkg"
	engine "github.com/nuts-foundation/nuts-go-core"
	"github.com/spf13/pflag"
)

// NewEventOctopusEngine creates the engine configuration for nuts-go.
func NewEventOctopusEngine() *engine.Engine {
	i := pkg.EventOctopusInstance()

	return &engine.Engine{
		Name:      "Events octopus",
		Config:    &i.Config,
		ConfigKey: "events",
		Configure: i.Configure,
		FlagSet:   flagSet(),
		Routes: func(router runtime.EchoRouter) {
			api.RegisterHandlers(router, &api.Wrapper{Eo: i})
		},
		Start:    i.Start,
		Shutdown: i.Shutdown,
	}
}

func flagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("event octopus", pflag.ContinueOnError)

	flags.Int(pkg.ConfigRetryInterval, pkg.ConfigRetryIntervalDefault, "Retry delay in seconds for reconnecting")
	flags.Int(pkg.ConfigNatsPort, pkg.ConfigNatsPortDefault, "Port for Nats to bind on")
	flags.String(pkg.ConfigConnectionstring, pkg.ConfigConnectionStringDefault, "db connection string for event store")

	return flags
}
