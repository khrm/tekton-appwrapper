# Makefile for tekton-appwrapper controller

# Variables
REGISTRY ?= ko.local
IMAGE_NAME ?= tekton-appwrapper-controller
VERSION ?= v0.1.0
NAMESPACE ?= tekton-pipelines

# Go variables
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# ko variables
KO_DOCKER_REPO ?= $(REGISTRY)/$(IMAGE_NAME)

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the controller binary
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/controller ./cmd/controller/main.go

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linters
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy
	go mod verify

##@ Build and Deploy

.PHONY: ko-build
ko-build: ## Build container image with ko
	ko build --local ./cmd/controller/

.PHONY: ko-publish
ko-publish: ## Build and publish container image with ko
	ko build --push ./cmd/controller/

.PHONY: ko-apply
ko-apply: ## Build and apply manifests with ko
	ko apply -f config/

.PHONY: ko-resolve
ko-resolve: ## Resolve ko references in manifests
	ko resolve -f config/

.PHONY: deploy
deploy: ## Deploy the controller to cluster
	kubectl apply -f config/

.PHONY: undeploy
undeploy: ## Remove the controller from cluster
	kubectl delete -f config/

.PHONY: install-crds
install-crds: ## Install AppWrapper CRD
	kubectl apply -f https://raw.githubusercontent.com/project-codeflare/appwrapper/main/config/crd/bases/workload.codeflare.dev_appwrappers.yaml

.PHONY: uninstall-crds
uninstall-crds: ## Remove AppWrapper CRD
	kubectl delete -f https://raw.githubusercontent.com/project-codeflare/appwrapper/main/config/crd/bases/workload.codeflare.dev_appwrappers.yaml

##@ Development

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@echo "Installing AppWrapper CRD..."
	$(MAKE) install-crds
	@echo "Deploying controller..."
	$(MAKE) ko-apply

.PHONY: dev-cleanup
dev-cleanup: ## Cleanup development environment
	@echo "Cleaning up development environment..."
	$(MAKE) undeploy
	$(MAKE) uninstall-crds

.PHONY: logs
logs: ## Get controller logs
	kubectl logs -n $(NAMESPACE) -l app=tekton-appwrapper-controller -f

.PHONY: status
status: ## Check controller status
	kubectl get pods -n $(NAMESPACE) -l app=tekton-appwrapper-controller
	kubectl get appwrappers -A

##@ Release

.PHONY: release
release: ## Create a release
	@echo "Creating release $(VERSION)..."
	git tag $(VERSION)
	git push origin $(VERSION)

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
