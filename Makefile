lint: ## Run linters. Use make install-linters first.
	GO111MODULE=off vendorcheck ./...
	GO111MODULE=on golangci-lint run -c .golangci.yml ./...
	# The govet version in golangci-lint is out of date and has spurious warnings, run it separately
	GO111MODULE=on go vet -all ./...

install:
	GO111MODULE=on go build -o ~/.skycoin/bin/skywire-updater ./cmd/skywire-updater
	export PATH=$PATH:$HOME/.skycoin/bin
	skywire-updater init-config
	cp -R ./scripts ~/.skycoin/skywire-updater/scripts

install-linters: ## Install linters
	GO111MODULE=on go get -u github.com/FiloSottile/vendorcheck
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	GO111MODULE=on go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	# GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.16.0

format: ## Formats the code. Must have goimports installed (use make install-linters).
	GO111MODULE=on goimports -w -local github.com/watercompany/skywire-updater ./pkg
	GO111MODULE=on goimports -w -local github.com/watercompany/skywire-updater ./cmd

dep: ## sorts dependencies
	GO111MODULE=on go mod vendor -v

test: ## Run tests for net
	@mkdir -p coverage/
	GO111MODULE=on go test -race -tags no_ci -cover -timeout=5m ./pkg/...