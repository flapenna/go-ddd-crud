BUF_IMAGE = bufbuild/buf
MOCKERY_IMAGE = vektra/mockery
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

.PHONY: buf-migrate proto proto-lint clean build run mocks test coverage unit integration

buf-migrate:
	docker run --rm --volume "$(PWD):/workspace" --workdir /workspace $(BUF_IMAGE) config migrate

proto:
	docker run --rm --volume "$(PWD):/workspace" --workdir /workspace $(BUF_IMAGE) dep update
	docker run --rm --volume "$(PWD):/workspace" --workdir /workspace $(BUF_IMAGE) generate

clean:
	rm -f bin/go-ddd-crud

build: clean
	go build -o bin/go-ddd-crud ./cmd/server

run: build
	./bin/go-ddd-crud

mocks:
	docker run --rm --volume "$(PWD):/workspace" --workdir /workspace $(MOCKERY_IMAGE)

# Run tests and generate coverage report
test:
	go test -v --tags=unit,integration -coverprofile=$(COVERAGE_FILE) ./...

# Run unit tests
unit:
	go test -v --tags=unit -coverprofile=$(COVERAGE_FILE) ./...

# Run integration tests
integration:
	go test -v --tags=integration -coverprofile=$(COVERAGE_FILE) ./...

# Generate HTML coverage report
coverage:
	go tool cover -html=$(COVERAGE_FILE)
	@echo "Coverage report generated: $(COVERAGE_HTML)"