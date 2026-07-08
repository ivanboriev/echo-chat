BINARY_NAME=chat
BUILD_DIR=build

.PHONY: build server connect

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME)  main.go

server: build
	@./$(BUILD_DIR)/$(BINARY_NAME)


connect:
	telnet localhost 8080	