#!/bin/bash

# 检查Go是否已安装
if [ -d "/usr/local/go" ]; then
    echo "检测到Go已安装在/usr/local/go目录"
else
    echo "未检测到Go安装，请先安装Go"
    exit 1
fi

# 检测shell类型
SHELL_TYPE=$(basename "$SHELL")
echo "当前shell类型: $SHELL_TYPE"

# 配置环境变量文件路径
if [ "$SHELL_TYPE" = "zsh" ]; then
    PROFILE_FILE="$HOME/.zshrc"
    echo "将配置 .zshrc 文件"
elif [ "$SHELL_TYPE" = "bash" ]; then
    PROFILE_FILE="$HOME/.bash_profile"
    echo "将配置 .bash_profile 文件"
else
    echo "不支持的shell类型: $SHELL_TYPE"
    exit 1
fi

# 检查环境变量是否已配置
if grep -q "export PATH=\$PATH:/usr/local/go/bin" "$PROFILE_FILE"; then
    echo "Go环境变量已配置在 $PROFILE_FILE 中"
else
    # 添加Go环境变量到配置文件
    echo "\n# Go环境变量配置" >> "$PROFILE_FILE"
    echo "export PATH=\$PATH:/usr/local/go/bin" >> "$PROFILE_FILE"
    echo "export GOPATH=\$HOME/go" >> "$PROFILE_FILE"
    echo "export PATH=\$PATH:\$GOPATH/bin" >> "$PROFILE_FILE"
    echo "export GOPROXY=https://goproxy.cn" >> "$PROFILE_FILE"
    
    echo "Go环境变量已添加到 $PROFILE_FILE 中"
fi

# 应用环境变量
echo "正在应用环境变量..."
source "$PROFILE_FILE"

# 验证Go环境变量
echo "验证Go环境变量:"
which go
go version

echo "\nGo环境变量配置完成，重启终端或运行 'source $PROFILE_FILE' 使配置生效"
echo "配置已添加到 $PROFILE_FILE，将在每次开机时自动加载"


export GOPROXY=https://goproxy.cn
source "$PROFILE_FILE"