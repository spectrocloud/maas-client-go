# If you update this file, please follow:
# https://www.thapaliya.com/en/writings/well-documented-makefiles/

# Meta
.DEFAULT_GOAL:=help
COVER_DIR=_build/cov

##@ Help Targets
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[0m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Dev Targets
env:  ## Display GOENV
	go env

fmt:  ## Format your code
	go fmt  ./...

lint:  ## Lint your code
	golangci-lint run ./...

dev-lint:  ## Lint your code, dev-mode
	golangci-lint run ./... --timeout 10m '--tests=false' '--disable=unused'

test:  ## Run unit tests
	@mkdir -p $(COVER_DIR)
	rm -f $(COVER_DIR)/*.out
	go test -v -covermode=count -coverprofile=$(COVER_DIR)/pkg_unit.out ./...

vet:
	go vet ./...
