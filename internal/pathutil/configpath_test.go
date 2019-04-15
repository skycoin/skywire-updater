package pathutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// https://nathanleclaire.com/blog/2015/10/10/interfaces-and-composition-for-effective-unit-testing-in-golang/

var wd, err = os.Getwd()

type inputs = struct {
	args      []string
	argsIndex int
	env       string
	defaults  ConfigPaths
}

var tables = []struct {
	description string
	in          inputs
	expectedMsg string
}{
	{
		description: "args",
		in:          inputs{[]string{"/home/anto/.skycoin/skywire-updater/config.yml"}, 0, "SW_MANAGER_CONFIG", ConfigPaths{"HOME": filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml")}},
		expectedMsg: filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml"),
	},
	// {
	// 	description: "WD",
	// 	in:          inputs{[]string{}, 0, "SW_MANAGER_CONFIG", ConfigPaths{"WD": filepath.Join(wd, "config.yml")}},
	// 	expectedMsg: filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml"),
	// },
	// {
	// 	description: "HOME",
	// 	in:          inputs{[]string{}, 0, "SW_MANAGER_CONFIG", ConfigPaths{"HOME": filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml")}},
	// 	expectedMsg: filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml"),
	// },
}

func TestFindConfigPath(t *testing.T) {

	for _, tt := range tables {
		t.Run(tt.description, func(t *testing.T) {

			assert := assert.New(t)
			configPath := FindConfigPath(tt.in.args, tt.in.argsIndex, tt.in.env, tt.in.defaults)

			assert.Nil(err)
			assert.Equal(configPath, tt.expectedMsg, "They should be equal")
		})
	}
}
