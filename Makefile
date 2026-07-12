.PHONY: build test lint serve clean

WASM_OUT = docs/main.wasm
WASM_EXEC = docs/wasm_exec.js
GO_WASM_SRC = ./cmd/wasm/

build: clean
	@mkdir -p docs
	GOOS=js GOARCH=wasm go build -o $(WASM_OUT) $(GO_WASM_SRC)
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_EXEC)
	@[ -f web/index.html ] && cp web/index.html docs/ || true
	@[ -d web/css ] && find web/css -name '*.css' | head -1 > /dev/null 2>&1 && cp -r web/css docs/ || true
	@[ -d web/js ] && find web/js -name '*.js' | head -1 > /dev/null 2>&1 && cp -r web/js docs/ || true

test:
	@go test ./internal/... 2>&1 || { \
		if go test ./internal/... 2>&1 | grep -q "no packages to test"; then \
			echo "No test packages found yet."; \
		else \
			exit 1; \
		fi; \
	}

lint:
	GOOS=js GOARCH=wasm golangci-lint run ./...

serve: build
	@echo "Serving at http://localhost:8080"
	python3 -m http.server 8080 --directory docs

clean:
	rm -rf docs/*
