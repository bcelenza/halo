TARGET_DIR=build

.PHONY: clean prepare build release

clean:
	rm -rf $(TARGET_DIR)

prepare:
	mkdir -p $(TARGET_DIR)/bin

build: prepare
	go build -o $(TARGET_DIR)/bin/halo .

release: clean build
