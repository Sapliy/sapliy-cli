BINARY_NAME=sapliy
INSTALL_PATH=/usr/local/bin

.PHONY: all build install clean

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Sapliy CLI installed to $(INSTALL_PATH)/$(BINARY_NAME)"

clean:
	rm -f $(BINARY_NAME)
