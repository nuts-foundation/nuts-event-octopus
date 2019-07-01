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

package cmd

import (
	"github.com/nuts-foundation/nuts-event-octopus/engine"
	cfg "github.com/nuts-foundation/nuts-go/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"sync"
)
var e = engine.NewEventOctopusEngine()
var rootCmd = &cobra.Command{
	Short: "test command",
	Run: func(cmd *cobra.Command, args []string) {
		var endWaiter sync.WaitGroup
		endWaiter.Add(1)
		var signalChannel chan os.Signal
		signalChannel = make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt)
		go func() {
			<-signalChannel
			endWaiter.Done()
		}()
		endWaiter.Wait()
	},
}

func Execute() {
	c := cfg.NutsConfig()
	c.IgnoredPrefixes = append(c.IgnoredPrefixes, e.ConfigKey)
	cfg.RegisterEngine(e)
	c.RegisterFlags(rootCmd, e)

	if err := c.Load(rootCmd); err != nil {
		panic(err)
	}

	c.PrintConfig(logrus.StandardLogger())

	if err := c.InjectIntoEngine(e); err != nil {
		panic(err)
	}

	if err := e.Configure(); err != nil {
		panic(err)
	}

	if err := e.Start(); err != nil {
		panic(err)
	}

	rootCmd.Execute()
}
