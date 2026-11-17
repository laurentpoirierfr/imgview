BINARY  := imgview
PKG     := ./cmd/imgview
BIN_DIR := bin
PREFIX  ?= $(HOME)/.local

GO      ?= go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)

.PHONY: all build run install uninstall test vet fmt tidy clean snapshot release-check help

all: build

build: ## Compile le binaire dans bin/
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(BINARY) $(PKG)

run: ## Lance imgview (ARGS=chemin)
	$(GO) run $(PKG) $(ARGS)

install: build ## Installe dans $(PREFIX)/bin
	install -Dm755 $(BIN_DIR)/$(BINARY) $(PREFIX)/bin/$(BINARY)

uninstall: ## Supprime le binaire installé
	rm -f $(PREFIX)/bin/$(BINARY)

test: ## Lance les tests
	$(GO) test ./...

vet: ## Analyse statique
	$(GO) vet ./...

fmt: ## Formate le code
	gofmt -w .

tidy: ## Nettoie go.mod / go.sum
	$(GO) mod tidy

clean: ## Supprime bin/ et dist/
	rm -rf $(BIN_DIR) dist

snapshot: ## Build local de toutes les plateformes avec GoReleaser (dist/)
	goreleaser release --snapshot --clean

release-check: ## Vérifie .goreleaser.yml
	goreleaser check

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*## "}{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
