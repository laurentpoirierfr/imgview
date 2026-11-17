BINARY  := imgview
PKG     := ./cmd/imgview
BIN_DIR := bin
PREFIX  ?= $(HOME)/.local

GO      ?= go
LDFLAGS ?= -s -w

.PHONY: all build run install uninstall test vet fmt tidy clean help

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

clean: ## Supprime bin/
	rm -rf $(BIN_DIR)

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*## "}{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
