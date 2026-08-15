GOFMT_FILES := ./apps/... ./packages/...

.PHONY: fmt
fmt:
	gofmt -w $$(find apps packages -name '*.go' -type f 2>/dev/null)

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: check
check: lint test

.PHONY: server-dev
server-dev:
	air -c apps/server/air.toml
