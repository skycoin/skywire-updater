[![Build Status](https://travis-ci.com/skycoin/skywire-updater.svg?branch=mainnet)](https://travis-ci.com/skycoin/skywire-updater)

# Skywire Updater

`skywire-updater` is responsible for updating services associated with Skywire. Actions include checking for updates, and actually updating the service.

```
$ skywire-updater -h

skywire-updater is responsible for checking for updates, and updating services
associated with skywire. Services to be updated will be based on specified configuration files.

Usage:
  skywire-updater [command]

Available Commands:
  help        Help about any command
  init-config generates a configuration file
  update      update services based on configuration file

Flags:
  -h, --help   help for skywire-updater

Use "skywire-updater [command] --help" for more information about a command.
```

## Installation

These instructions details the installation and configuration of `skywire-updater` in a user's home directory.

Prerequisites:
- Have [golang](https://golang.org/dl/) installed. We need a version that supports [go modules](https://github.com/golang/go/wiki/Modules).

```bash
# Clone.
$ git clone https://github.com/watercompany/skywire-updater
$ cd ./skywire-updater

# Build.
$ GO111MODULE=on go build -o ~/.skycoin/bin/skywire-updater ./cmd/skywire-updater

# Export path (would be a good idea to add this line to ~/.profile).
$ export PATH=$PATH:$HOME/.skycoin/bin

# Generate default config to ~/.skycoin/skywire-updater/config.yml
$ skywire-updater init-config

# Copy scripts.
$ cp -R ./scripts ~/.skycoin/skywire-updater/scripts

# Run.
$ skywire-updater

```

## Configuration

A configuration file contains the following sections:
- `paths:` - Specifies paths for the `skywire-updater`.
- `interfaces:` - Specifies network interface settings.
- `services.defaults:` - Specifies default values for `services`.
- `services.services:` - Specifies services that the `skywire-updater` is responsible for.

Here is an example configuration with comments:

```yaml
paths: # Configures paths.
  db-file: "/usr/local/skywire-updater/db.json"      # Database file location ("/usr/local/skywire-updater/db.json" if unspecified).
  scripts-path: "/usr/local/skywire-updater/scripts" # Scripts folder location ("/usr/local/skywire-updater/scripts" if unspecified).

interfaces: # Configures network interfaces.
  addr: ":8080"     # Address to bind and listen from (":7280" if unspecified).
  enable-rest: true # Whether to enable RESTful interface served from {addr}/api/ (true if unspecified).
  enable-rpc: true  # Whether to enable RPC interface served from {addr}/rpc/ (true if unspecified).


services: # Configures services.
  defaults: # Configures default field values.
    main-branch: "master"     # Default 'main-branch' field value.
    bin-dir: "/usr/local/bin" # Default bin directory filed value.
    interpreter: "/bin/sh"    # Default 'interpreter' field value.
    envs:                     # Default 'envs' field values.
      - "APP_DIR=/usr/local/skywire/apps/bin"
  services:
    skywire: # Service name/ID. This service is named "skywire".
      repo:         "github.com/skycoin/skywire" # Repository URL. Should be of format: <domain>/<owner>/<name> . Will be saved in SWU_REPO env for scripts.
      main-branch:  "stable"                     # Main branch's name. Default will be used if not set. Will be saved in SWU_MAIN_BRANCH env for scripts.
      bin-dir:      "/usr/local/skycoin/bin"     # Bin Directory to build into. Will be saved in SWU_BIN_DIR for scripts.
      main-process: "skywire-node"               # Main executable's name. Will be saved in SWU_MAIN_PROCESS env for scripts.
      checker:                                            # Defines the service's checker (used to check for available updates).
        type: "script"                                    # Type of checker. Valid: "script"(default), "github_release".
        script: "check/bin-diff"                          # Required if checker type is "script": Specifies script to run (within '--scripts-dir' arg).
        interpreter: "/bin/bash"                          # Required if checker type is "script": Specifies script interpreter. Default will be used if not set.
        args: - "-v"                                      # Optional: Additional arguments for checker scripts.
        envs:                                             # Optional: Set environment variables that can be used by checker.
          - "APP_DIR=/usr/local/skywire/default-apps/bin" # This overrides default's APP_DIR definition.
      updater:                                            # Defines the service's updater (actually updates the service's binaries and relevant files).
        type: "script"                                    # Type of updater. Only "script"(default) is supported.
        script: "update/skywire"                          # Required if updater type is "script": Specifies script to run (within '--scripts-dir' arg).
        interpreter: "/bin/bash"                          # Required if updater type is "script": Specifies script interpreter. Default will be used if not set.
        args: - "-v"                                      # Optional: Additional arguments for updater scripts.
        envs:                                             # Optional: Set environment variables that can be used by updater.
          - "APP_DIR=/usr/local/skywire/default-apps/bin" # This overrides default's APP_DIR definition.

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

import "github.com/skycoin/skywire-updater/pkg/api"

func main() {
	client, err := api.DialRPC(":7280")
	// interact with client ...
}
```

Note that the RPC and REST interfaces of the `skywire-updater` are served on the same port (but on different paths).