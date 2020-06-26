package main

import (
	"github.com/nuts-foundation/nuts-event-octopus/engine"
	"github.com/nuts-foundation/nuts-go-core/docs"
)

func main() {
	docs.GenerateConfigOptionsDocs("README_options.rst", engine.NewEventOctopusEngine().FlagSet)
}
