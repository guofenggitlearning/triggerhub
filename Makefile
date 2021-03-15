SHELL := /bin/bash
PATH  := $(PATH):$(shell go env GOPATH)/bin
.DEFAULT_GOAL := help

PROJECT_NAME=$(shell basename "$(PWD)")
PROTOC := bin/protoc
SOURCES=$(wildcard proto/*.proto)
TARGETS=$(patsubst %.proto, %.pb.go, $(SOURCES))


define install_protoc
	@case "$$(uname)" in \
		linux|Linux) \
			curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip > protoc.zip \
			;;\
		darwin|Darwin) \
			curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-osx-x86_64.zip > protoc.zip \
			;;\
		*) \
			echo "Unsupported platform: $$(uname)" ;\
			exit 1 ;\
	esac ;\
	rm -Rf ./bin
	unzip -d . protoc.zip ;\
	rm -R protoc.zip readme.txt include
endef

define install_protoc_go
	if [ ! -d "$$GOPATH" ] ; then \
		export GOPATH="$$HOME/go" ;\
	fi ; \
	go get google.golang.org/protobuf/cmd/protoc-gen-go ; \
	go install google.golang.org/protobuf/cmd/protoc-gen-go ; \
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc ; \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
endef

#-----------------------------------------------------------------------
# HELP
#-----------------------------------------------------------------------

## help: Display this message

.PHONY: help
help:
	@echo
	@echo " Available actions in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' Makefile | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## :

## init: Install external dependencies
init: $(PROTOC) protoc-go-plugin

## clean: Remove the build artifacts
clean:
	rm -Rf build proto/*.go

## :

#-----------------------------------------------------------------------
# RECIPES
#-----------------------------------------------------------------------


## all: Generate the source code for all supported languages
all: $(TARGETS)

$(TARGETS): $(PROTOC) $(SOURCES)
	rm -rf $@
	mkdir -p proto
	$(PROTOC) -I=$(PWD)/proto \
		--go_out=$(PWD)/proto --go_opt=paths=source_relative \
		--go-grpc_out=$(PWD)/proto --go-grpc_opt=paths=source_relative \
		$(patsubst %, $(PWD)/%, $(SOURCES))

#-----------------------------------------------------------------------
# COMPILERS
#-----------------------------------------------------------------------

$(PROTOC):
	$(call install_protoc)
	@echo "Using protoc from $(PROTOC)"

.PHONY: protoc-go-plugin
protoc-go-plugin:
	$(call install_protoc_go)
