package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/urfave/cli"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/api"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/memory"
	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/sql"
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
		cli.StringFlag{Name: "db", Value: "user=postgres database=transports disablessl=true", Usage: "Postgres connection string for the transport database"},
	},
	Action: func(c *cli.Context) error {
		tdb, err := sql.NewStore(c.String("db"))
		if err != nil {
			return err
		}
		defer tdb.Close()
		if err := tdb.Migrate(context.Background()); err != nil {
			return err
		}

		ndb := memory.NewStore()

		s := struct {
			store.TransportStore
			store.NonceStore
		}{tdb, ndb}

		l, err := net.Listen("tcp", c.String("bind"))
		if err != nil {
			return err
		}
		log.Printf("Listening on %s", l.Addr().String())

		return http.Serve(l, api.New(s, api.APIOptions{}))
	},
}
