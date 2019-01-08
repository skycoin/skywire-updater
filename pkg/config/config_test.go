package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/watercompany/skywire-updater/pkg/config"
)

func TestServices(t *testing.T) {
	var expectedServiceMaps = map[string]config.ServiceConfig{
		"skywire-manager": {
			OfficialName:               "skywire-manager",
			LocalName:                  "manager",
			UpdateScript:               "skywire.sh",
			UpdateScriptInterpreter:    "/bin/bash",
			UpdateScriptTimeout:        "20m",
			UpdateScriptExtraArguments: []string{"-web-dir ${GOPATH}/pkg/github.com/skycoin/skywire/static/skywire-manager"},
			CheckScript:                "generic-service-check-update.sh",
			CheckScriptInterpreter:     "/bin/bash",
			CheckScriptTimeout:         "20m",
			ActiveUpdateChecker:        "simple",
			Repository:                 "/skycoin/skywire",
			Updater:                    "custom",
		},
		"skywire-node": {
			OfficialName:            "skywire-node",
			LocalName:               "node",
			UpdateScript:            "skywire.sh",
			UpdateScriptInterpreter: "/bin/bash",
			UpdateScriptTimeout:     "20m",
			UpdateScriptExtraArguments: []string{
				"-connect-manager -manager-address 127.0.0.1:5998",
				"-manager-web 127.0.0.1:8000",
				"-discovery-address discovery.skycoin.net:5999-034b1cd4ebad163e457fb805b3ba43779958bba49f2c5e1e8b062482904bacdb68",
				"-address :5000",
				"-web-port :6001",
			},
			ActiveUpdateChecker:    "simple",
			CheckScript:            "generic-service-check-update.sh",
			CheckScriptInterpreter: "/bin/bash",
			CheckScriptTimeout:     "20m",
			Repository:             "/skycoin/skywire",
			Updater:                "custom",
		},
	}

	c := config.NewFromFile("../../configuration.skywire.yml")

	assert.Equal(t, expectedServiceMaps, c.Services)
}

func TestUpdaters(t *testing.T) {
	var expectedUpdaters = map[string]config.UpdaterConfig{
		"custom": {
			Kind: "custom",
		},
	}

	c := config.NewFromFile("../../configuration.skywire.yml")

	assert.Equal(t, expectedUpdaters, c.Updaters)
}

func TestActiveUpdateChekers(t *testing.T) {
	var expectedActiveUpdateChekcers = map[string]config.FetcherConfig{
		"simple": {
			Kind:      "simple",
			NotifyURL: "http://localhost:8989/update",
		},
	}

	c := config.NewFromFile("../../configuration.skywire.yml")

	assert.Equal(t, expectedActiveUpdateChekcers, c.ActiveUpdateCheckers)
}
