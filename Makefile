 # 获取版本信息
VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go 构建标志
LDFLAGS := -X github.com/xinkaiwang/hermes/internal/common.Version=$(VERSION) -X github.com/xinkaiwang/hermes/internal/common.GitCommit=$(GIT_COMMIT) -X github.com/xinkaiwang/hermes/internal/common.BuildTime=$(BUILD_TIME)
GOFLAGS := -ldflags "$(LDFLAGS)"

# Docker 相关变量
DOCKER_REPO := xinkaiw
DOCKER_IMAGE := hermes
DOCKER_TAG := v$(VERSION)

.PHONY: all hello test clean run

all: hello hermes

hello:
	@echo "Building hello..."
	@mkdir -p bin
	go build $(GOFLAGS) -o bin/hello ./cmd/hello

hermes:
	@echo "Building hello..."
	@mkdir -p bin
	go build $(GOFLAGS) -o bin/hermes ./service/hermes

# Docker 相关目标
docker-build:
	@echo "Building Docker image $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) -f Dockerfile .

docker-push:
	@echo "Pushing Docker image $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker push $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/

run: hello
	@echo "Running hello..."
	./bin/hello
