APP = bac
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...

BIN ?= $(OUTPUT_DIR)/$(APP)
KUBECTL_BIN ?= $(OUTPUT_DIR)/kubectl-$(APP)

GO_FLAGS ?= -mod=vendor
GO_TEST_FLAGS ?= -race -cover

GO_PATH ?= $(shell go env GOPATH)
GO_CACHE ?= $(shell go env GOCACHE)

INSTALL_LOCATION ?= /usr/local/bin

ARGS ?=

.EXPORT_ALL_VARIABLES:

.PHONY: $(BIN)
$(BIN):
	go build $(GO_FLAGS) -o $(BIN) $(CMD)

build: $(BIN)

install: build
	install -m 0755 $(BIN) $(INSTALL_LOCATION)

# creates a kubectl prefixed binary, "kubectl-$APP", and when installed under $PATH, will be
# visible as "kubectl $APP".
# See https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/
# Not employing krew at this time
.PHONY: kubectl
kubectl: BIN = $(KUBECTL_BIN)
kubectl: $(BIN)

kubectl-install: BIN = $(KUBECTL_BIN)
kubectl-install: kubectl install

clean:
	rm -rf "$(OUTPUT_DIR)"

# runs all tests
test: test-unit

.PHONY: test-unit
test-unit:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(CMD) $(PKG) $(ARGS)

