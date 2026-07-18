# Makefile — gmh (git-meta-harness) CLI
# Builda binário estático para linux/darwin/windows × amd64/arm64
#
# Uso:
#   make build         # builda para o OS atual
#   make all           # builda para todos OS/arch
#   make test          # roda os testes
#   make install       # instala em $GOPATH/bin
#   make clean         # remove binários

BINARY     := gmh
VERSION    := $(shell cat ../VERSION 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE       := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

PLATFORMS  := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: help build all test install clean

help: ## Mostra ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Builda para o OS/arch atual
	@mkdir -p bin
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY) .

all: ## Builda para todos OS/arch (linux/darwin/windows × amd64/arm64)
	@mkdir -p bin
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		output=bin/$(BINARY)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then output=$$output.exe; fi; \
		echo "==> Building $$os/$$arch -> $$output"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build $(LDFLAGS) -o $$output . || exit 1; \
	done

test: ## Roda os testes
	go test -race -shuffle=on ./...

install: build ## Instala em $GOPATH/bin
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

clean: ## Remove binários
	rm -rf bin/
