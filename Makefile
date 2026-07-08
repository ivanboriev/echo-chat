BINARY_NAME=chat
BUILD_DIR=build

.PHONY: build server connect clean fmt lint

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME)  ./cmd/chat

server: build
	@./$(BUILD_DIR)/$(BINARY_NAME)


connect:
	telnet localhost 8080

clean:
	@rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

lint:
	go vet ./...