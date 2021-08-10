PROJECT = dtrack
BIN = dtrack
LOCALDIR = local
DOCKER_COMPOSE = ./$(LOCALDIR)/docker-compose.yaml
RESULTSDIR = test/results
LINTCMD = $(TEMPDIR)/golangci-lint run --tests=false --config .golangci.yaml
BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)

$(TEMPDIR):
	mkdir -p $(TEMPDIR)

## Variable assertions

## Tasks
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'

.PHONY: bootstrap-dtrack
bootstrap-dtrack: $(TEMPDIR) ## Download dependancy track latest docker-compose - see .tmp
	$(call title,Bootstrapping dependancy track)
	@curl -L -o $(TEMPDIR)/docker-compose.yaml https://dependencytrack.org/docker-compose.yml 
	@sudo apt install ansible

# .PHONY: bootstrap-go
# bootstrap-go:
# 	go mod download

# .PHONY: bootstrap
# bootstrap: $(RESULTSDIR) bootstrap-dtrack ## Download and install all go dependencies (+ prep tooling in the ./tmp dir)
# 	$(call title,Bootstrapping dependencies)

.PHONY: start-local
start-local:  ## Deploy local API, Frontend services (docker-compose)
	@docker-compose -p ${PROJECT} -f $(DOCKER_COMPOSE) up  -d 
	@make setup-users

.PHONY: stop-local
stop-local:  ## Stop local API, Frontend services (docker-compose)
	@docker-compose -p ${PROJECT} -f $(DOCKER_COMPOSE) stop

.PHONY: restart-local ## Restart local API, Frontend services (docker-compose)
restart-local: stop-local start-local

.PHONY: down-local
down-local:  ## Down local API, Frontend services (docker-compose)
	@docker-compose -p ${PROJECT} -f $(DOCKER_COMPOSE) down

.PHONY: info-local
info-local: ## Display services info (docker-compose)
	@docker-compose -p ${PROJECT} -f $(DOCKER_COMPOSE) ps

.PHONY: clean-local
clean-local: ## Clean local API, Frontend services (docker-compose)
	@docker-compose -p ${PROJECT} -f $(DOCKER_COMPOSE) down -v

.PHONY: attach-local
attach-local: ## Attach to api image - drop to shell (docker-compose)
	@docker exec -it ${PROJECT}_dtrack-apiserver_1 sh

.PHONY: attach-log-local
attach-log-local: ## Attach to api log (docker-compose)
	@docker attach -f $(DOCKER_COMPOSE) ${PROJECT}_dtrack-apiserver_1

.PHONY: tail
tail-local: ## Tail service logs (docker-compose)
	@docker-compose -p ${PROJECT}  -f $(DOCKER_COMPOSE) logs -f

.PHONY: tail
setup-users: ## Setup initial users and teams
	@ansible-playbook ansible/setup_users.yaml --ask-vault-pass

.PHONY: integration
integration: ## Run integration tests
	$(call title,Running integration tests)
	go test -v ./test/integration

# .PHONY: build
# build: $(SNAPSHOTDIR) ## Build release snapshot binaries and packages

# $(SNAPSHOTDIR): ## Build snapshot release binaries and packages
# 	$(call title,Building snapshot artifacts)
# 	# create a config with the dist dir overridden
# 	echo "dist: $(SNAPSHOTDIR)" > $(TEMPDIR)/goreleaser.yaml
# 	cat .goreleaser.yaml >> $(TEMPDIR)/goreleaser.yaml

# 	# build release snapshots
# 	BUILD_GIT_TREE_STATE=$(GITTREESTATE) \
# 	$(TEMPDIR)/goreleaser release --debug --skip-publish --skip-sign --rm-dist --snapshot --config $(TEMPDIR)/goreleaser.yaml