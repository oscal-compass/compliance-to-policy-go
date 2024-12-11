GO_BUILD_PACKAGES := ./cmd/...
GO_BUILD_BINDIR :=./bin

all: vendor test-unit
.PHONY: all


build: prep-build-dir
	go build -o ${GO_BUILD_BINDIR} ${GO_BUILD_PACKAGES}
.PHONY: build

prep-build-dir:
	mkdir -p ${GO_BUILD_BINDIR}
.PHONY: prep-build-dir

clean:
	@rm -rf ./$(GO_BUILD_BINDIR)/*
.PHONY: clean

vendor:
	go mod tidy
	go mod verify
	go mod vendor
.PHONY: vendor

test-unit:
	go test -coverprofile=coverage.out -race -count=1 ./...
.PHONY: test-unit

sanity: vendor format vet
	git diff --exit-code
.PHONY: sanity

format:
	go fmt ./...
.PHONY: format

vet:
	go vet ./...
.PHONY: vet

generate-protobuf:
	protoc api/proto/*.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_out=. --go_opt=paths=source_relative --proto_path=.
.PHONY: generate-protobuf
