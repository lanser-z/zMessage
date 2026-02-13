# Makefile for zMessage Server

# Go 环境变量
GOROOT := /usr/local/go
GOPATH := $(HOME)/go
PATH := $(GOROOT)/bin:$(GOPATH)/bin:$(PATH)
export GOROOT GOPATH PATH

# Go 命令
GO := go
GOFLAGS := -v

.PHONY: all test build clean deps fmt lint vet

# 默认目标
all: deps test

# 下载依赖
deps:
	$(GO) mod download
	$(GO) mod verify

# 格式化代码
fmt:
	$(GO) fmt ./...

# 静态检查
vet:
	$(GO) vet ./...

# 运行测试
test:
	$(GO) test -v -race -cover ./...

# 运行特定包的测试
test-user:
	$(GO) test -v -race -cover ./modules/user

test-dal:
	$(GO) test -v -race -cover ./dal

# 构建服务端
build:
	$(GO) build -o ../bin/server ./cmd/server

# 清理
clean:
	$(GO) clean
	rm -f ../bin/server

# 安装开发工具
install-tools:
	$(GO) install golang.org/x/tools/cmd/goimports@latest
