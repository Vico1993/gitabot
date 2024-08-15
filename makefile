.PHONY: ensure_deps build test lint watch

default: test lint

ensure_deps:
	go mod vendor
	go mod tidy

build:
	@ echo "🛠  Start building 🛠"
	@ go build -a \
			 -o "bin/bot" "./internal"
	@ echo "🛠  Build done 🛠"

test:
	go test -v -mod=vendor ./...

lint:
	@ echo "🪛  Start linting 🪛"
	@ golangci-lint run ./...
	@ echo "🪛  Lint done 🪛"

lint_fix:
	@ echo "🪛  Start linting with Fix 🪛"
	@ golangci-lint run --fix  ./...
	@ echo "🪛  Fixed your lint 🪛"

watch:
	@ echo "👀  Continue working... I'm watching 👀"
	@ gow -c run ./internal