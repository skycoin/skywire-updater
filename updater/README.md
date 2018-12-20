
![skycoin_logo](https://user-images.githubusercontent.com/26845312/32426705-d95cb988-c281-11e7-9463-a3fce8076a72.png)

# Updater

updater is a service that can be configured to look for updates for other services, and when those are
available, send a POST request to a configured address notifying of such update. After that the API
exposed by updater can be used to update any service.

updater allows to configure:

1. **How to look for updates:** Currently support to perform periodical checks on github or dockerhub (we call this active update checkers), there is also a naive active checker that will just call the updater every interval, or to subscribe to a publisher that will send updater update notifications (we call this passive update checkers), the only publisher supported so far is [nats](https://nats.io/).
    
2. **How to update services:** We call this Updaters, and currently there are only supported custom scripts.
    
3. **Which services to update:** A list of services with a per-service configuration to allow updater to know where to look for updates and how to update it.

# Installation

Have [golang](https://golang.org/dl/) installed.

Have [dep](https://golang.github.io/dep/docs/installation.html) installed.
Optional, if you want to regenerate project dependencies. It's optional since due
to the size of some of the dependencies I have choose to upload the vendor folder, which
holds that dependencies but with the unused files purged, so it requires much less space:

1. `go get github.com/watercompany/skywire-services`
2. `cd $GOPATH/src/github.com/watercompany/skywire-services/updater`
3. `dep ensure (only needed if you want to regenerate the dependencies)`
4. `go install ./cmd/...`

After that you should have an `updater` binary in your `$GOPATH` variable. If `$GOPATH` is included in your `$PATH` variable you should have the command available to use from terminal.

Now, from the root directory of updater you can run `make install`, which will create an `skywire-updater`
dir in your home dir and place updater scripts inside, which will make the provided configuration files
to work by default.

# Usage

```bash
Usage of updater:
  -api-port string
    	port in which to listen (default "8080")
  -config string
    	yaml configuration file (default "../configuration.yml")
```

You will need to use one of the provided configuration files or use your own.
`configuration.messaging.yml` is provided to update messaging binaries.
`configuration.skywire.yml` is provided to update skywire binaries.

You can run `updater -config ./configuration.messaging.yml` to look for `messaging`
binaries updates and notify http://localhost:8989/update with a POST request when one
is available. You can change this address in the configuration file in
active_update_checkers.naive.notify_url.

After you are notified with the update you can request `updater` to update one of the
services, for example:
`curl "localhost:8080/update/skywire-messaging-client"`

# API

#### Update
Update updates the given service, which should match the name given in the configuration
```
URI: /update/:service_name
Method: POST
```

#### Register
Register registers a new service into updater in order to look for new versions of it. A default checker
using `generic-service-check-update.sh` is used and will be updated using `generic-service.sh`
```
URI: /register/:service_name
Method: POST
Content-Type: application/json
Body: {
 	"notify-url":"<url where to send a POST request upon new version available>",
 	"current-version":"<current version of the service to register>",
 	"repository":"<repository where to check for updates>"
}
```

#### Unregister
Unregister unregisters a service from updater, which will stop looking for new versions
```
URI: /register/:service_name
Method: POST
```

# Configuration

A `configuration.example.yml` file is provided as a reference file.

The configuration file has 3 sections:

1. **active_update_checkers:**
    Is a list containing named active checkers configurations, each named item should hold the next configuration:

```
notify_url: url to send a post request to when an update has been found. Example: "http://localhost:8989/update"
interval: interval in which to check if there is a new update. Example: "30s"
kind: which kind of checker is it, git or dockerhub: "git"
```

2. **passive_update_checkers:**
    Is a list containing named passive checkers configurations, each named item should hold the next configuration:

```
message-broker: kind of the message-broker to subscribe to. Example: "nats"
topic: topic on the message-broker to subscribe to. Example: "top"
urls: urls of message-broker cluster to join to. One or more. Example:
    - "http://localhost:4222"
```

3. **updaters:**
    Is a list containing named updaters, each named item should hold the next configuration:

```
kind: the kind of updater to instantiate, swarm or custom. Example: "custom"
```

4. **services:**
    Is a list containing named services, each named item should hold the next configuration:

```
official_name: official name for the service, this is how is it named on their documentation, etc. Example "maria"
local_name: name of the service running on your machine, in case it's different. This is needed for swarm or for the scripts to find the service they need to update. Example: "mariadb"
check_script: name of the check update script in case you are using a custom updater. Example:  "custom_script.sh"
check_script_interpreter: interpreter of the script in case you are using a custom updater. Example: "/bin/bash"
check_script_timeout: timeout for the script in case you are using a custom updater. Example: "20s"
check_script_extra_arguments: extra arguments that you want to be passed to the updater script in case you are using a custom updater. Example: ["-a 1"]
update_script: name of the update script in case you are using a custom updater. Example:  "custom_script.sh"
update_script_interpreter: interpreter of the script in case you are using a custom updater. Example: "/bin/bash"
update_script_timeout: timeout for the script in case you are using a custom updater. Example: "20s"
update_script_extra_arguments: extra arguments that you want to be passed to the updater script in case you are using a custom updater. Example: ["-a 1"]
active_update_checker: name of a previously defined active update checker if used. Example: "dockerhub_fetcher"
active_update_checker: name of a previously defined passive update checker if used. Example: "nats"
repository: repository of the service in the format /:owner/:name for lookup on dockerhub or github. Example: "/library/mariadb"
check_tag: name of the tag to check for updates, this is used to check for updates on github or dockerhub. Don't needed for passive checkers. Example: "latest"
updater: previously defined updater configuration name. Example: "custom"
```

# Running on Docker
Firstable, you can pull the image: `docker pull skycoin/updater` or you can build it yourself: `docker build . [your-image-name]`.

By default, the docker image doesn't provides any configuration, so you will have to place one inside using the tag `-v [your-configuration]:/updater/configuration.yml`, updater is looking for that file by default, if you use a different one you should tell updater with `-config [path-to-your-config]` flag.

If you want to use the Docker image to update other services running under docker swarm, or maybe even under the docker daemon, you will also need to mount the docker daemon socket inside, again using the tag `-v /var/run/docker.sock:/var/run/docker.sock`.

An example of command to run the docker image would be:
`docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/configuration.yml:/updater/configuration.yml --rm -it skycoin/updater:0.0.1`.
