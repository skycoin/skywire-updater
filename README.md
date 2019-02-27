
![skycoin_logo](https://user-images.githubusercontent.com/26845312/32426705-d95cb988-c281-11e7-9463-a3fce8076a72.png)

# Skywire Updater

`skywire-updater` is responsible for updating services associated with Skywire. Actions include checking for updates, and actually updating the service.

```
Updates skywire services

Usage:
  skywire-updater [flags]

Flags:
      --config-file string   path to updater's configuration file (default "/Users/anonymous/go/src/github.com/watercompany/skywire-updater/config.skywire.yml")
      --db-file string       path to db file (creates if not exist) (default "/Users/anonymous/.skywire/updater/db.json")
  -h, --help                 help for skywire-updater
      --http-addr string     address in which to serve http api (disabled if not set) (default ":6781")
      --rpc-addr string      address in which to serve rpc api (disabled if not set) (default ":6782")
      --scripts-dir string   path to dir containing scripts (default "/Users/anonymous/go/src/github.com/watercompany/skywire-updater/scripts")
```

## Installation

Prerequisites:
- Have [golang](https://golang.org/dl/) installed.
- Enable [go modules](https://github.com/golang/go/wiki/Modules).

```bash
# Download.
$ go get -d -u github.com/watercompany/skywire-updater

# Install.
$ cd ${GOPATH}/src/github.com/watercompany/skywire-updater
$ go install ./...
```

## RESTful API

### List services
```
Method: GET
URI: /services
```

### Check for updates for given service
```
Method: GET
URI: /services/:service_name/check
```

### Update given service
```
Method: POST
URI: /services/:service_name/update/:version
```

# Configuration

The `config.example.yml` file is the default configuration.

The configuration file contains a list of `services`.

# Running on Docker
The image can be pulled via `docker pull skycoin/updater` or it can be built with `docker build . [your-image-name]`.

By default, the docker image doesn't provides any configuration, so you will have to place one inside using the tag `-v [your-configuration]:/updater/configuration.yml`, updater is looking for that file by default, if you use a different one you should tell updater with `-config [path-to-your-config]` flag.

An example of command to run the docker image would be:
`docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/configuration.yml:/updater/configuration.yml --rm -it skycoin/updater:0.0.1`.
