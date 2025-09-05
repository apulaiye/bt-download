# Go 示例项目

这是一个基本的 Go 语言项目示例，展示了 Go 项目的标准结构和简单功能实现。

## 项目结构

```
.
├── cmd/          # 应用程序入口
│   └── main.go   # 主程序
├── pkg/          # 可重用的库代码
│   └── greeting/ # 问候功能包
├── internal/     # 私有应用和库代码
├── go.mod        # Go 模块定义
└── README.md     # 项目说明
```

## 功能

这个示例项目实现了一个简单的问候功能，通过 `greeting` 包提供 `Hello` 函数。

## 使用方法

### 安装

```bash
git clone https://github.com/tiny/repo.git
cd repo
```

### 运行

```bash
go run cmd/main.go
```

### 测试

```bash
go test ./...
```

## 许可证

MIT