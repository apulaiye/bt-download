# 可执行文件名称
BINARY_NAME := bt-download
# 入口文件路径
ENTRY_POINT := ./cmd/main.go
# 版本信息（可通过git commit hash自动获取）
VERSION := $(shell git rev-parse --short HEAD)
# 构建时间
BUILD_TIME := $(shell date +"%Y-%m-%d %H:%M:%S")

# 默认目标：编译项目
all: build

# 编译项目（使用vendor依赖，添加版本信息）
build:
	@echo "开始编译 $(BINARY_NAME)..."
	go build \
		-ldflags "-X 'main.version=$(VERSION)' -X 'main.buildTime=$(BUILD_TIME)'" \
		-o $(BINARY_NAME) $(ENTRY_POINT)
	@echo "编译完成: $(BINARY_NAME)"

# 编译并运行
run: build
	@echo "运行 $(BINARY_NAME)..."
	./$(BINARY_NAME)

# 清理编译产物
clean:
	@echo "清理编译产物..."
	rm -f $(BINARY_NAME)
	rm -rf ./dist  # 清理跨平台编译目录
	@echo "清理完成"

# 运行测试
test:
	@echo "运行测试..."
	go test  -v ./...

# 跨平台编译（生成Linux和macOS版本）
cross-build:
	@echo "开始跨平台编译..."
	mkdir -p dist
	# Linux 64位
	GOOS=linux GOARCH=amd64 go build  -o dist/$(BINARY_NAME)-linux-amd64 $(ENTRY_POINT)
	# macOS 64位
	GOOS=darwin GOARCH=amd64 go build  -o dist/$(BINARY_NAME)-darwin-amd64 $(ENTRY_POINT)
	@echo "跨平台编译完成，输出目录: dist/"

# 显示帮助信息
help:
	@echo "可用命令:"
	@echo "  make          - 编译项目（默认目标）"
	@echo "  make build    - 编译项目"
	@echo "  make run      - 编译并运行项目"
	@echo "  make clean    - 清理编译产物"
	@echo "  make test     - 运行测试"
	@echo "  make cross-build - 跨平台编译（Linux/macOS）"
	@echo "  make help     - 显示帮助信息"
