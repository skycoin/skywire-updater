.DEFAULT_GOAL := help
.PHONY : check lint install-linters dep test 

OPTS=GO111MODULE=on 

check: lint test ## Run linters and tests

lint: ## Run linters. Use make install-linters first.
	GO111MODULE=off vendorcheck ./...
	${OPTS} golangci-lint run -c .golangci.yml ./...
	# The govet version in golangci-lint is out of date and has spurious warnings, run it separately
	${OPTS} go vet -all ./...

install: dep
	${OPTS} go build -o ~/.skycoin/bin/skywire-updater ./cmd/skywire-updater
	export PATH=$PATH:$HOME/.skycoin/bin
	skywire-updater init-config
	cp -R ./scripts ~/.skycoin/skywire-updater/scripts

install-linters: ## Install linters
	${OPTS} go get -u github.com/FiloSottile/vendorcheck
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	${OPTS} go get -u golang.org/x/tools/cmd/goimports 

format: ## Formats the code. Must have goimports installed (use make install-linters).
	${OPTS} goimports -w -local github.com/skycoin/skywire-updater ./pkg
	${OPTS} goimports -w -local github.com/skycoin/skywire-updater ./cmd

dep: ## sorts dependencies
	${OPTS} go mod vendor -v

test: ## Run tests for net
	@mkdir -p coverage/
	${OPTS} go test -race -tags no_ci -cover -timeout=5m ./pkg/...

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
