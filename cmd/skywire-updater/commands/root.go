package commands

import (
	"os"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/spf13/cobra"
)

var log = logging.MustGetLogger("skywire-updater")

// RootCmd is the command to run when no sub-commands are specified.
var RootCmd = &cobra.Command{
	Use:   "skywire-updater [command]",
	Short: "Updates skywire services",
	Long: `
skywire-updater is responsible for checking for updates, and updating services
associated with skywire. Services to be updated will be based on a specified configuration file.`,
}

// Execute executes root CLI command and add subcommands.
func Execute() {

	RootCmd.AddCommand(initConfigCmd)
	RootCmd.AddCommand(updateCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
