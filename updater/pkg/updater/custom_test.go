package updater_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/watercompany/skywire-services/updater/config"
	"github.com/watercompany/skywire-services/updater/pkg/logger"
	"github.com/watercompany/skywire-services/updater/pkg/updater"
)

const testScript = `
#!/bin/bash

echo "service {$1}"
echo "version {$2}"
shift 2

echo "arguments {$@}"
`

func TestCustom(t *testing.T) {
	customConfig := &config.Configuration{
		Updaters: map[string]config.UpdaterConfig{
			"test": {
				Kind: "custom",
			},
		},
		Services: map[string]config.ServiceConfig{
			"myservice": {
				LocalName:            "myservice",
				OfficialName:         "myservice",
				UpdateScriptInterpreter:    "/bin/bash",
				UpdateScript:         "-s",
				UpdateScriptExtraArguments: []string{"<<<", testScript, "arg2"},
				UpdateScriptTimeout:        "5s",
				Updater:              "test",
			},
		},
	}
	customUpdater := updater.New("custom", customConfig)

	log := logger.NewLogger("myservice")
	err := <-customUpdater.Update("myservice", "thisversion", log)

	assert.NoError(t, err)
}

func TestTimeout(t *testing.T) {
	customConfig := &config.Configuration{
		Updaters: map[string]config.UpdaterConfig{
			"test": {
				Kind: "custom",
			},
		},
		Services: map[string]config.ServiceConfig{
			"myservice": {
				LocalName:            "myservice",
				OfficialName:         "myservice",
				Updater:              "test",
				UpdateScriptInterpreter:    "top",
				UpdateScript:         "",
				UpdateScriptExtraArguments: []string{},
				UpdateScriptTimeout:        "1s",
			},
		},
	}
	customUpdater := updater.New("custom", customConfig)

	log := logger.NewLogger("myservice")
	err := <-customUpdater.Update("myservice", "thisversion", log)

	assert.Error(t, err)
}
