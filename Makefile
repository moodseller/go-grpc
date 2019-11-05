CI_PIPELINE_ID?=local
VERSION:=${CI_PIPELINE_ID}

build-protoc:
	@docker build -t go-testing-proto -q -f ./tools/protoc/Dockerfile .

protoc: build-protoc
	@docker run --rm \
		-v $$(pwd)/api:/app/api \
		go-testing-proto

run-server:
	@go build -o ./bin/server ./cmd/server/.
	@./bin/server \
		-env development \

build:
	@go build \
    ./pkg/... \
	./cmd/server/...
