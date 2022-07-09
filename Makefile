OUT_DIR = out
SERVER_DIR = server
BIN_NAME = ts3-bot
GOFLAGS = -race
DOCKER = docker

help: ## available make tasks
	@awk 'BEGIN {FS = ":.*##"; printf "targets:\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  %-16s%s\n", $$1, $$2 } /^##@/ { printf "\n%s\n" } ' $(MAKEFILE_LIST)

clean: ## clean project
	-rm -rf $(OUT_DIR)
	-rm -rf $(SERVER_DIR)
	go clean
	go mod tidy

assets: ## bundle web assets
	-mkdir -p assets
	go-assets-builder assets -o assets.go # https://github.com/nothub/go-assets-builder

lint: assets ## lint golang code
	go vet

build: assets ## build binary artifact
	go build $(GOFLAGS) -o $(OUT_DIR)/$(BIN_NAME)

run: build ## run bot
	./$(OUT_DIR)/$(BIN_NAME)

server: ## run ts3 server
	$(DOCKER) pull n0thub/ts3
	$(DOCKER) run                   \
	  --name=ts3server              \
	  --interactive                 \
	  --tty                         \
	  --rm                          \
	  -e TS3SERVER_LICENSE=accept   \
	  -p 9987:9987/udp              \
	  -p 30033:30033                \
	  -p 127.0.0.1:10080:10080      \
	  -v ${PWD}/$(SERVER_DIR):/data \
	  n0thub/ts3                    \
	  query_protocols=http

.PHONY: help clean assets lint build run server
