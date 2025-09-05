#!/bin/bash

echo "开始下载Go安装包..."
echo "请确保您的网络连接正常"

# 设置变量
GO_VERSION="1.21.0"
DOWNLOAD_URL="https://golang.google.cn/dl/go${GO_VERSION}.darwin-amd64.pkg"
PKG_FILE="go${GO_VERSION}.darwin-amd64.pkg"

# 下载Go安装包
echo "从国内镜像下载Go ${GO_VERSION}..."
curl -L -o "${PKG_FILE}" "${DOWNLOAD_URL}"

if [ $? -ne 0 ]; then
    echo "下载失败，尝试从备用镜像下载..."
    BACKUP_URL="https://studygolang.com/dl/golang/go${GO_VERSION}.darwin-amd64.pkg"
    curl -L -o "${PKG_FILE}" "${BACKUP_URL}"
    
    if [ $? -ne 0 ]; then
        echo "下载失败，请检查网络连接或手动下载安装。"
        echo "手动下载地址：https://studygolang.com/dl"
        exit 1
    fi
fi

echo "下载完成，开始安装..."
echo "请输入您的密码以安装Go"

# 安装Go
sudo installer -pkg "${PKG_FILE}" -target /

if [ $? -ne 0 ]; then
    echo "安装失败，请尝试手动安装下载的安装包：${PKG_FILE}"
    exit 1
fi

# 清理下载文件
rm "${PKG_FILE}"

# 配置环境变量
echo "配置Go环境..."

# 设置GOPATH
GOPATH="$HOME/go"
mkdir -p "$GOPATH"

# 配置国内镜像
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

echo "Go安装和配置完成！"
echo "Go版本："
go version

echo "\n您现在可以运行Go项目了："
echo "cd /Users/tiny/repo"
echo "go run cmd/main.go"