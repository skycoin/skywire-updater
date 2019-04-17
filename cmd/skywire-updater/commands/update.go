package commands

import (
	"fmt"
	"net"
	"net/http"

	"github.com/skycoin/skywire-updater/internal/pathutil"
	"github.com/skycoin/skywire-updater/pkg/api"
	"github.com/skycoin/skywire-updater/pkg/store"
	"github.com/skycoin/skywire-updater/pkg/update"
	"github.com/spf13/cobra"
)

const configEnv = "SW_UPDATER_CONFIG"

var defaultPaths = pathutil.UpdaterDefaults()

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update services based on configuration file",
	Long: fmt.Sprintf(`
	update takes one optional argument [path] which specifies the path to the
	configuration file to use. If no [path] is specified, the following
	directories are searched in order:

	  1. %s
	  2. %s
	  3. %s`, defaultPaths[pathutil.WorkingDirLoc], defaultPaths[pathutil.HomeLoc], defaultPaths[pathutil.LocalLoc]),
	Run: func(_ *cobra.Command, args []string) {
		configPath := pathutil.FindConfigPath(args, 0, configEnv, pathutil.UpdaterDefaults())

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
