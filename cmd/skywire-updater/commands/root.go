package commands

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/spf13/cobra"

	"github.com/watercompany/skywire-updater/pkg/api"
	"github.com/watercompany/skywire-updater/pkg/store"
	"github.com/watercompany/skywire-updater/pkg/update"
)

var (
	configPath string
	dbPath     string
	scriptsDir string
	httpAddr   string
	rpcAddr    string

	log = logging.MustGetLogger("skywire-updater")
)

func init() {
	// defaults.
	configPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/watercompany/skywire-updater/config.skywire.yml")
	dbPath     = filepath.Join(os.Getenv("HOME"), ".skywire/updater/db.json")
	scriptsDir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/watercompany/skywire-updater/scripts")
	httpAddr   = ":6781"
	rpcAddr    = ":6782"

	// flags.
	rootCmd.PersistentFlags().StringVar(&configPath, "config-file", configPath, "path to updater's configuration file")
	rootCmd.Flags().StringVar(&dbPath, "db-file", dbPath, "path to db file (creates if not exist)")
	rootCmd.Flags().StringVar(&scriptsDir, "scripts-dir", scriptsDir, "path to dir containing scripts")
	rootCmd.Flags().StringVar(&httpAddr, "http-addr", httpAddr, "address in which to serve http api (disabled if not set)")
	rootCmd.Flags().StringVar(&rpcAddr, "rpc-addr", rpcAddr, "address in which to serve rpc api (disabled if not set)")
}

var rootCmd = &cobra.Command{
	Use:   "skywire-updater",
	Short: "Updates skywire services",
	PreRun: func(_ *cobra.Command, _ []string) {
		checkEnv := func(key string) {
			if _, ok := os.LookupEnv(key); !ok {
				log.Fatalf("%s needs to be set", key)
			}
		}
		for _, key := range []string{"HOME", "GOPATH"} {
			checkEnv(key)
		}
	},
	Run: func(_ *cobra.Command, _ []string) {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer close(sigCh)

		db, err := store.NewJSON(dbPath)
		if err != nil {
			log.WithError(err).Fatalln("failed to load db")
			return
		}

		conf, err := update.NewConfig(configPath)
		if err != nil {
			log.WithError(err).Fatalln("failed to load config")
			return
		}

		srv := update.NewManager(db, scriptsDir, conf)

		if httpAddr != "" {
			l, err := net.Listen("tcp", httpAddr)
			if err != nil {
				log.WithError(err).Fatalln("failed to listen http")
				return
			}
			log.Infof("http listening on %s", l.Addr())
			go func() {
				if err := http.Serve(l, api.HandleHTTP(srv)); err != nil {
					log.WithError(err).Fatalln("failed to serve http")
					return
				}
			}()
		}
		if rpcAddr != "" {
			l, err := net.Listen("tcp", rpcAddr)
			if err != nil {
				log.WithError(err).Fatalln("failed to listen rpc")
				return
			}
			log.Infof("rpc listening on %s", l.Addr())
			go func() {
				if err := http.Serve(l, api.HandleRPC(srv)); err != nil {
					log.WithError(err).Fatalln("failed to serve http")
					return
				}
			}()
		}

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
