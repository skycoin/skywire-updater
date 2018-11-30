package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/watercompany/skywire-services/updater/config"
	"github.com/watercompany/skywire-services/updater/pkg/api"
)

var configFile string

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.StringVar(&configFile, "config", "./configuration.yml", "yaml configuration file")
	flag.Parse()

	configuration := config.NewFromFile(configFile)
 	gateway := api.NewServerGateway(configuration)
 	go gateway.Start("localhost:8080")
	<-sigs
	gateway.Stop()
}
