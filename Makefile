# Variables
BINARY_NAME=go-deploy
BUILD_DIR=bin
MAIN_FILE=main.go
BUILDTIMESTAMP=$(shell date -u +%Y%m%d%H%M%S)
EXT=$(if $(filter windows,$(GOOS)),.exe,)

# Targets
.PHONY: all clean build run clean docs release test acc e2e lint

all: build

build:
	@echo "Building the application..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -ldflags="-X github.com/NVIDIA/k8s-dra-driver-gpu/internal/info.version=v1.34.2" -o $(BUILD_DIR)/$(BINARY_NAME)$(EXT) .
	@echo "Build complete."

run: build
	@echo "Running the application..."
	@./$(BUILD_DIR)/$(BINARY_NAME)$(EXT)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

docs:
	@cd scripts && ./generate-docs.sh && ./generate-types.sh

release: docs
	@echo "Building the application..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -mod=readonly -ldflags "-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)$(EXT) .
	@echo "Build complete."

test: clean build
	@go test ./test/acc/...
	@echo "Starting go-deploy in test mode..."
	@./$(BUILD_DIR)/$(BINARY_NAME)$(EXT) --mode=test & echo $$! > go_deploy.pid
	@echo "Waiting for API to become ready..."
	@until curl --output /dev/null --silent --head --fail http://localhost:8080/healthz; do \
		echo "Waiting for API to start"; \
		sleep 1; \
	done
	@echo "API is ready!"
	@echo "Running e2e tests..."
	@go test ./test/e2e/...
	@if [ -f go_deploy.pid ]; then \
		echo "Stopping go-deploy..."; \
		kill $$(cat go_deploy.pid) && rm -f go_deploy.pid; \
	fi

acc: clean build
	@go test ./test/acc/...

e2e: clean build
	@echo "Starting go-deploy in test mode..."
	@./$(BUILD_DIR)/$(BINARY_NAME)$(EXT) --mode=test & echo $$! > go_deploy.pid
	@echo "Waiting for API to become ready..."
	@until curl --output /dev/null --silent --head --fail http://localhost:8080/healthz; do \
		echo "Waiting for API to start"; \
		sleep 1; \
	done
	@echo "API is ready!"
	@echo "Running e2e tests..."
	@go test ./test/e2e/...
	@if [ -f go_deploy.pid ]; then \
		echo "Stopping go-deploy..."; \
		kill $$(cat go_deploy.pid) && rm -f go_deploy.pid; \
	fi

lint: 
	@./scripts/check-lint.sh
