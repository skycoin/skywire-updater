package commands

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/spf13/cobra"

	"github.com/skycoin/skywire/pkg/util/pathutil"

	"github.com/skycoin/skywire-updater/pkg/api"
	"github.com/skycoin/skywire-updater/pkg/store"
	"github.com/skycoin/skywire-updater/pkg/update"
)

const configEnv = "SW_UPDATER_CONFIG"

var log = logging.MustGetLogger("skywire-updater")

// UpdaterDefaults returns the default config paths for Skywire-Updater.
func UpdaterDefaults() pathutil.ConfigPaths {
	paths := make(pathutil.ConfigPaths)
	if wd, err := os.Getwd(); err == nil {
		paths[pathutil.WorkingDirLoc] = filepath.Join(wd, "config.yml")
	}
	paths[pathutil.HomeLoc] = filepath.Join(pathutil.HomeDir(), ".skycoin/skywire-updater/config.yml")
	paths[pathutil.LocalLoc] = "/usr/local/skycoin/skywire-updater/config.yml"
	return paths
}

var defaultPaths = UpdaterDefaults()

// RootCmd is the command to run when no sub-commands are specified.
// TraverseChildren set to true enables cobra to parse local flags on each command before executing target command
var RootCmd = &cobra.Command{
	Use:   "skywire-updater [config-path]",
	Short: "Updates skywire services",
	Long: fmt.Sprintf(`
skywire-updater is responsible for checking for updates, and updating services
associated with skywire.

It takes one optional argument [config-path] which specifies the path to the
configuration file to use. If no [config-path] is specified, the following
directories are searched in order:

  1. %s
  2. %s
  3. %s`, defaultPaths[pathutil.WorkingDirLoc], defaultPaths[pathutil.HomeLoc], defaultPaths[pathutil.LocalLoc]),
	TraverseChildren: true,
	Run: func(_ *cobra.Command, args []string) {

		configPath := pathutil.FindConfigPath(args, 0, configEnv, defaultPaths)

		log.Infof("config path: '%s'", configPath)
		conf := update.NewConfig(".", "./bin")
		if err := conf.Parse(configPath); err != nil {
			log.WithError(err).Fatalln("failed to load config")
			return
		}

		log.Infof("db path: '%s'", conf.Paths.DBFile)
		db, err := store.NewJSON(conf.Paths.DBFile)
		if err != nil {
			log.WithError(err).Fatalln("failed to load db")
			return
		}

		srv := update.NewManager(db, conf)

		l, err := net.Listen("tcp", conf.Interfaces.Addr)
		if err != nil {
			log.WithError(err).Fatalln("failed to listen http")
			return
		}

		log.Infof("serving on address '%s'", l.Addr())
		if err := http.Serve(l, api.Handle(srv, conf.Interfaces.EnableREST, conf.Interfaces.EnableRPC)); err != nil {
			log.WithError(err).Fatalln("failed to serve http")
			return
		}

		if err := srv.Close(); err != nil {
			log.WithError(err).Error()
		}
	},
}

// Execute executes root CLI command and add subcommands.
func Execute() {

	RootCmd.AddCommand(initConfigCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
