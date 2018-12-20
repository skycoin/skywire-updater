lint: ## Run linters. Use make install-linters first.
	vendorcheck ./...
	golangci-lint run -c .golangci.yml ./...
	# The govet version in golangci-lint is out of date and has spurious warnings, run it separately
	go vet -all ./...

install:
	mkdir -p ${HOME}/.skywire-updater
	cp -r ./scripts ${HOME}/.skywire-updater/scripts

install-linters: ## Install linters
	go get -u github.com/FiloSottile/vendorcheck
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

format: ## Formats the code. Must have goimports installed (use make install-linters).
	goimports -w -local github.com/watercompany/skywire-services/autoupdater ./pkg
	goimports -w -local github.com/watercompany/skywire-services/autoupdater ./cmd

test: ## Run tests for net
	@mkdir -p coverage/
	go test -coverpkg="github.com/watercompany/skywire-services/autoupdater/..." -coverprofile=coverage/go-test-cmd.coverage.out -timeout=5m ./pkg/...
