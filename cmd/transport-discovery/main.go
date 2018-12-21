package main

import (
	"log"
	"net"
	"net/http"

	"github.com/urfave/cli"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/api"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
)

func main() {
	app := cli.NewApp()
	app.Name = "skywire transport-discovery"
	app.Commands = []cli.Command{
		serve,
	}

	app.RunAndExitOnError()
}

var serve = cli.Command{
	Name:  "serve",
	Usage: "Starts the server",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "bind", Value: ":8080", Usage: "Where to bind to"},
		cli.StringFlag{Name: "db", Value: "redis://localhost:6379", Usage: "Redis URL"},
	},
	Action: func(c *cli.Context) error {
		s, err := store.New("redis", c.String("db"))
		if err != nil {
			return err
		}

		l, err := net.Listen("tcp", c.String("bind"))
		if err != nil {
			return err
		}
		log.Printf("Listening on %s", l.Addr().String())

		return http.Serve(l, api.New(s, api.APIOptions{}))
	},
}
