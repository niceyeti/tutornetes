BIN_DIR = bin
PROTO_DIR = proto
SERVER_DIR = server
CLIENT_DIR = client

SHELL := bash
SHELL_VERSION = $(shell echo $$BASH_VERSION)
UNAME := $(shell uname -s)
OS = linux $(shell uname -rm)
PACKAGE = $(shell head -1 go.mod | awk '{print $$2}')
CHECK_DIR_CMD = test -d $@ || (echo "\033[31mDirectory $@ doesn't exist\033[0m" && false)
HELP_CMD = grep -E '^[a-zA-Z_-]+:.*?\#\# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?\#\# "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
RM_F_CMD = rm -f
RM_RF_CMD = ${RM_F_CMD} -r
SERVER_BIN = ${SERVER_DIR}
CLIENT_BIN = ${CLIENT_DIR}

.DEFAULT_GOAL := help
.PHONY := all help clean crud clean_crud test bump about

all: rebuild ## Generate Pbs and build

rebuild: clean bump crud ## Rebuild the whole project

clean: clean_crud ## Clean generated files
	${RM_F_CMD} ssl/*.crt
	${RM_F_CMD} ssl/*.csr
	${RM_F_CMD} ssl/*.key
	${RM_F_CMD} ssl/*.pem
	${RM_RF_CMD} ${BIN_DIR}

clean_crud: ## Clean generated files for CRUD app
	${RM_F_CMD} ./${PROTO_DIR}/*.pb.go

cproto: ## Compile proto definitions. This must be done or the generated go package will not be found by tools like `go get -u ./...`
	protoc -I${PROTO_DIR} --go_opt=module=${PACKAGE} --go_out=. --go-grpc_opt=module=${PACKAGE} --go-grpc_out=. ${PROTO_DIR}/*.proto

## crud: $@ ## Generate Pbs and build for crud service
crud: cproto go.mod go.sum $(ls ${SERVER_DIR})
	go build -o ./${BIN_DIR}/${SERVER_BIN} ./${SERVER_DIR}
	go build -o ./${BIN_DIR}/${CLIENT_BIN} ./${CLIENT_DIR}

test: ## Launch tests
	go test ./...

bump: cproto go.mod go.sum ## Update packages version
	go get -u ./...
	@echo && echo "  ^---Ignoring github.com/golang/protobuf deprecation warning. TODO: address the dep warning." && echo

about: ## Display info related to the build
	@echo "OS: ${OS}"
	@echo "Shell: ${SHELL} ${SHELL_VERSION}"
	@echo "Protoc version: $(shell protoc --version)"
	@echo "Go version: $(shell go version)"
	@echo "Go package: ${PACKAGE}"
	@echo "Openssl version: $(shell openssl version)"

help: ## Show this help
	@${HELP_CMD}