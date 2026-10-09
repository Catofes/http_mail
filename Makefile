GO ?= go
BINARY ?= http_mail
DIST_DIR ?= dist
IMAGE ?= http_mail:local
BUILD_FLAGS := -trimpath -ldflags="-s -w"

.PHONY: build check integration release docker-build docker-check

build:
	CGO_ENABLED=0 $(GO) build $(BUILD_FLAGS) -o "$(BINARY)" .

check:
	git diff --check
	test -z "$$(gofmt -l *.go)"
	$(GO) vet ./...
	$(GO) test -race ./...

integration:
	bash scripts/test-postgres.sh

release:
	mkdir -p "$(DIST_DIR)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o "$(DIST_DIR)/http_mail-linux-amd64" .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o "$(DIST_DIR)/http_mail-linux-arm64" .
	cp config.json.example README.md "$(DIST_DIR)/"
	cd "$(DIST_DIR)" && sha256sum http_mail-linux-amd64 http_mail-linux-arm64 config.json.example README.md > SHA256SUMS

docker-build:
	docker build --tag "$(IMAGE)" .

# Check an existing local image or registry digest without rebuilding it.
docker-check:
	docker run --rm "$(IMAGE)" -h
