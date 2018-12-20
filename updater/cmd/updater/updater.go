package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/watercompany/skywire-services/updater/pkg/api"
	"github.com/watercompany/skywire-services/updater/pkg/config"
)

var (
	configFile string
	apiPort    string
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.StringVar(&configFile, "config", "./configuration.yml", "yaml configuration file")
	flag.StringVar(&apiPort, "api-port", "8080", "port in which to listen")
	flag.Parse()

	configuration := config.NewFromFile(configFile)
	gateway := api.NewServerGateway(configuration)
	go func() {
		err := gateway.Start("localhost:" + apiPort)
		if err != nil {
			panic(err)
		}
	}()

	<-sigs

	err := gateway.Stop()
	if err != nil {
		panic(err)
	}
}
