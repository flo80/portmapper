# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

PROTOC=protoc

# Binary path and names
BINARY_PATH=bin
BINARY_NAME_CLI=portmapper
BINARY_NAME_SERVER=portmapserver
BINARY_NAME_CLIENT=portmapclient

# Source path
SOURCE_CLI=cli
SOURCE_SERVER=cmd/server
SOURCE_CLIENT=cmd/client

# Build flags
BUILD_VERSION=`git rev-parse --short HEAD`
LDFLAGS=-ldflags "-X main.Build=$(BUILD_VERSION) -s -w"
LINUX=CGO_ENABLED=0 GOOS=linux GOARCH=amd64

.PHONY: deps clean proto

all: deps build 

build: proto
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME_CLI) ./$(SOURCE_CLI)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME_SERVER) ./$(SOURCE_SERVER)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME_CLIENT) ./$(SOURCE_CLIENT)


build_linux: proto
	$(LINUX) $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME_CLI)_linux ./$(SOURCE_CLI)
	$(LINUX) $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME_SERVER)_linux ./$(SOURCE_SERVER)
	$(LINUX) $(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH)/$(BINARY_NAME_CLIENT)_linux ./$(SOURCE_CLIENT)

deps:
	$(GOGET) "github.com/jackpal/gateway"
	$(GOGET) "github.com/jackpal/go-nat-pmp"
	$(GOGET) "github.com/metricube/upnp"
	$(GOGET) "golang.org/x/net/context"
	$(GOGET) "google.golang.org/grpc"
	$(GOGET) "github.com/golang/protobuf/proto"
	$(GOGET) "google.golang.org/genproto/googleapis/api/annotations"
	
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)/$(BINARY_NAME_CLI)
	rm -f $(BINARY_PATH)/$(BINARY_NAME_CLI)_linux
	rm -f $(BINARY_PATH)/$(BINARY_NAME_CLIENT)
	rm -f $(BINARY_PATH)/$(BINARY_NAME_CLIENT)_linux
	rm -f $(BINARY_PATH)/$(BINARY_NAME_SERVER)
	rm -f $(BINARY_PATH)/$(BINARY_NAME_SERVER)_linux

proto:
	$(PROTOC) -I api -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:api api/portmap.proto