package engine

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEventOctopusEngine_Start(t *testing.T) {
	t.Run("start in CLI mode", func(t *testing.T) {
		os.Setenv("NUTS_MODE", core.GlobalCLIMode)
		defer os.Unsetenv("NUTS_MODE")
		if !assert.NoError(t, core.NutsConfig().Load(&cobra.Command{})) {
			return
		}
		assert.NoError(t, NewEventOctopusEngine().Start())
	})
}

func TestEventOctopusEngine_Configure(t *testing.T) {
	t.Run("configure in CLI mode", func(t *testing.T) {
		os.Setenv("NUTS_MODE", core.GlobalCLIMode)
		defer os.Unsetenv("NUTS_MODE")
		if !assert.NoError(t, core.NutsConfig().Load(&cobra.Command{})) {
			return
		}
		assert.NoError(t, NewEventOctopusEngine().Configure())
	})
}