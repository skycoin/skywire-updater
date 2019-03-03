package commands

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/spf13/cobra"

	"github.com/watercompany/skywire-updater/pkg/api"
	"github.com/watercompany/skywire-updater/pkg/store"
	"github.com/watercompany/skywire-updater/pkg/update"
)

const defaultConfigPath = "/usr/local/skywire-updater/config.yml"

var log = logging.MustGetLogger("skywire-updater")

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("skywire-updater [%s]", defaultConfigPath),
	Short: "Updates skywire services",
	Run: func(_ *cobra.Command, args []string) {
		configPath := defaultConfigPath
		if len(args) > 0 {
			configPath = args[0]
		}

		log.Infof("config path: '%s'", configPath)
		conf, err := update.ParseConfig(configPath)
		if err != nil {
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

		if conf.Interfaces.EnableREST || conf.Interfaces.EnableRPC {
			l, err := net.Listen("tcp", conf.Interfaces.Addr)
			if err != nil {
				log.WithError(err).Fatalln("failed to listen http")
				return
			}
			log.Infof("serving on address '%s'", l.Addr())
			go func() {
				if err := http.Serve(l, api.Handle(srv, conf.Interfaces.EnableREST, conf.Interfaces.EnableRPC)); err != nil {
					log.WithError(err).Fatalln("failed to serve http")
					return
				}
			}()
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer close(sigCh)
		log.Infof("exited with sig: %s", <-sigCh)

		if err := srv.Close(); err != nil {
			log.WithError(err).Error()
		}
	},
}

// Execute executes root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal()
	}
}
