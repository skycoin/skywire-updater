[![Build Status](https://travis-ci.com/watercompany/skywire-updater.svg?token=U4rdXdKvUqSqMgvR66wF&branch=master)](https://travis-ci.com/watercompany/skywire-updater)

# Skywire Updater

`skywire-updater` is responsible for updating services associated with Skywire. Actions include checking for updates, and actually updating the service.

```
$ skywire-updater -h

Updates skywire services

Usage:
  skywire-updater [assets/config.default.yml] [flags]

Flags:
  -h, --help   help for skywire-updater
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

# Run.
$ skywire-updater
```

## Configuration

The `config.skywire.yml` file is the default configuration for skywire.

The configuration file contains the following sections:
- `paths:` - Specifies paths for the `skywire-updater`.
- `interfaces:` - Specifies network interface settings.
- `defaults:` - Specifies default values for `services`.
- `services:` - Specifies services that the `skywire-updater` is responsible for.

Here is an example configuration with comments:

```yaml
paths: # Configures paths.
  db-file: "/usr/local/skywire-updater/db.json"      # Database file location ("db.json" if unspecified).
  scripts-path: "/usr/local/skywire-updater/scripts" # Scripts folder location ("scripts" if unspecified).

interfaces: # Configures network interfaces.
  addr: ":8080" # Address to bind and listen from (":7280" if unspecified).
  enable-rest: true # Whether to enable RESTful interface served from {addr}/api/ (true if unspecified).
  enable-rpc: true  # Whether to enable RPC interface served from {addr}/rpc/ (true if unspecified).

defaults: # Configures default field values.
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

## RESTful Endpoints

- **List services**
    ```
    GET /api/services
    ```

- **Check for updates for given service**
    ```
    GET /api/services/:service_name/check
    ```

- **Update given service**
    ```
    POST /api/services/:service_name/update/:version
    ```

## RPC Endpoints

An RPC Client is provided in [/pkg/api/rpc.go](/pkg/api/rpc.go).

```go
package main

import "github.com/watercompany/skywire-updater/pkg/api"

func main() {
	client, err := api.DialRPC(":7280")
	// interact with client ...
}
```

Note that the RPC and REST interfaces of the `skywire-updater` are served on the same port (but on different paths).

## Running in Docker 
***WARNING: untested - requires updating.***

The image can be pulled via `docker pull skycoin/updater` or it can be built with `docker build . [your-image-name]`.

By default, the docker image doesn't provides any configuration, so you will have to place one inside using the tag `-v [your-configuration]:/updater/configuration.yml`, updater is looking for that file by default, if you use a different one you should tell updater with `-config [path-to-your-config]` flag.

An example of command to run the docker image would be:
`docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/configuration.yml:/updater/configuration.yml --rm -it skycoin/updater:0.0.1`.
