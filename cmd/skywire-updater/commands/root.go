package commands

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/spf13/cobra"

	"github.com/skycoin/skywire-updater/internal/pathutil"
	"github.com/skycoin/skywire-updater/pkg/api"
	"github.com/skycoin/skywire-updater/pkg/store"
	"github.com/skycoin/skywire-updater/pkg/update"
)

var log = logging.MustGetLogger("skywire-updater")

var defaultConfigPaths = [2]string{
	filepath.Join(pathutil.HomeDir(), ".skycoin/skywire-updater/config.yml"),
}

func findConfigPath() (string, error) {
	log.Info("configuration file is not explicitly specified, attempting to find one in default paths ...")
	for i, cPath := range defaultConfigPaths {
		if _, err := os.Stat(cPath); err != nil {
			log.Infof("- [%d/%d] '%s' does not exist", i, len(defaultConfigPaths), cPath)
		} else {
			log.Infof("- [%d/%d] '%s' exists (using this one)", i, len(defaultConfigPaths), cPath)
			return cPath, nil
		}
	}
	return "", errors.New("no configuration file found")
}

// RootCmd is the command to run when no sub-commands are specified.
var RootCmd = &cobra.Command{
	Use: "skywire-updater [config-path]",
	Long: fmt.Sprintf(`
skywire-updater is responsible for checking for updates, and updating services
associated with skywire.

It takes one optional argument [config-path] which specifies the path to the
configuration file to use. If no [config-path] is specified, the following 
directories are searched in order:

  1. %s
  2. %s`, defaultConfigPaths[0], defaultConfigPaths[1]),
	Short: "Updates skywire services",
	Run: func(_ *cobra.Command, args []string) {
		var configPath string
		if len(args) == 0 {
			var err error
			if configPath, err = findConfigPath(); err != nil {
				log.WithError(err).Fatal()
			}
		} else {
			configPath = args[0]
		}

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

// Execute executes root CLI command.
func Execute() {
	RootCmd.AddCommand(initConfigCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
