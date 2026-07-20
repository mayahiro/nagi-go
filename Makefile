.PHONY: build check format format-check lint test unicode

build:
	GOWORK=off GOTOOLCHAIN=local go build ./...

test:
	GOWORK=off GOTOOLCHAIN=local go test ./...

lint:
	GOWORK=off GOTOOLCHAIN=local go vet ./...

format:
	GOWORK=off GOTOOLCHAIN=local go -C tools tool goimports -local github.com/mayahiro/nagi-go -w ..

format-check:
	@unformatted="$$(find . -type f -name '*.go' -not -path './.git/*' -exec gofmt -l {} +)"; test -z "$$unformatted" || { printf '%s\n' "$$unformatted"; exit 1; }

check:
	$(MAKE) format-check
	$(MAKE) build
	$(MAKE) test
	$(MAKE) lint

unicode:
	@test -n "$(UNICODE_DATA)" || { echo "UNICODE_DATA must name the Unicode 17.0.0 source directory" >&2; exit 2; }
	GOTOOLCHAIN=local go run ./cmd/unicodegen -data-dir "$(UNICODE_DATA)" -out text/generated.go
