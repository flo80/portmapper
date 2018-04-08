# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary path and names
BINARY_PATH=bin
BINARY_NAME=portmapper

# Build flags
BUILD_VERSION=`git rev-parse --short HEAD`
LDFLAGS=-ldflags "-X main.Build=$(BUILD_VERSION) -s -w"

.PHONY: deps clean

all: deps build build_for_docker

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME) .

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME)_linux .

deps:
	$(GOGET) "github.com/jackpal/gateway"
	$(GOGET) "github.com/jackpal/go-nat-pmp"
	$(GOGET) "github.com/metricube/upnp"
	
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)/$(BINARY_NAME)
	rm -f $(BINARY_PATH)/$(BINARY_NAME)_linux
