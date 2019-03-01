[![Build Status](https://travis-ci.com/watercompany/skywire-updater.svg?token=U4rdXdKvUqSqMgvR66wF&branch=master)](https://travis-ci.com/watercompany/skywire-updater)

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

## Configuration

The `config.skywire.yml` file is the default configuration for skywire.

The configuration file contains two main sections:
- `default:` - Specifies default values for other fields.
- `services:` - Specifies services that the `skywire-updater` is responsible for.

Here is an example configuration with comments:

```yaml
default: # Configures default field values.
  main-branch: "master"  # Default 'main-branch' field value ("master" if unspecified).
  interpreter: "/bin/sh" # Default 'interpreter' field value ("/bin/bash" if unspecified).
  envs:                  # Default 'envs' field values (none if unspecified).
    - "BIN_DIR=/usr/local/skywire/bin"
    - "APP_DIR=/usr/local/skywire/bin/apps"

services: # Configures services.
  skywire: # Service name/ID. This service is named "skywire".
    repo:         "github.com/skycoin/skywire" # Repository URL. Should be of format: <domain>/<owner>/<name> . Will be saved in SKYUPD_REPO env for scripts.
    main-branch:  "stable"                     # Main branch's name. Default will be used if not set. Will be saved in SKYUPD_MAIN_BRANCH env for scripts.
    main-process: "skywire-node"               # Main executable's name. Will be saved in SKYUPD_MAIN_PROCESS env for scripts.
    checker:                                            # Defines the service's checker (used to check for available updates).
      type: "script"                                    # Type of checker. Valid: "script"(default), "github_release".
      script: "check/bin_diff"                          # Required if checker type is "script": Specifies script to run (within '--scripts-dir' arg).
      interpreter: "/bin/bash"                          # Required if checker type is "script": Specifies script interpreter. Default will be used if not set.
      args: - "-v"                                      # Optional: Additional arguments for checker scripts.
      envs:                                             # Optional: Set environment variables that can be used by checker.
        - "APP_DIR=/usr/local/skywire/bin/default-apps" # This overrides default's APP_DIR definition.
    updater:                                            # Defines the service's updater (actually updates the service's binaries and relevant files).
      type: "script"                                    # Type of updater. Only "script"(default) is supported.
      script: "update/skywire"                          # Required if updater type is "script": Specifies script to run (within '--scripts-dir' arg).
      interpreter: "/bin/bash"                          # Required if updater type is "script": Specifies script interpreter. Default will be used if not set.
      args: - "-v"                                      # Optional: Additional arguments for updater scripts.
      envs:                                             # Optional: Set environment variables that can be used by updater.
        - "APP_DIR=/usr/local/skywire/bin/default-apps" # This overrides default's APP_DIR definition.

  another-service: # Another service. This service is named "another-service".
    # The config for 'another-service' goes here ...
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

## Running in Docker 
***WARNING: untested - requires updating.***

The image can be pulled via `docker pull skycoin/updater` or it can be built with `docker build . [your-image-name]`.

By default, the docker image doesn't provides any configuration, so you will have to place one inside using the tag `-v [your-configuration]:/updater/configuration.yml`, updater is looking for that file by default, if you use a different one you should tell updater with `-config [path-to-your-config]` flag.

An example of command to run the docker image would be:
`docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/configuration.yml:/updater/configuration.yml --rm -it skycoin/updater:0.0.1`.
