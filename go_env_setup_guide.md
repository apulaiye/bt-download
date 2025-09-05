# Go环境变量配置指南

本指南将帮助您配置Go语言的环境变量，并设置为开机自动加载。

## 自动配置方法

我们提供了一个自动配置脚本 `setup_go_env.sh`，您可以直接运行此脚本来配置Go环境变量：

```bash
# 使脚本可执行
chmod +x setup_go_env.sh

# 运行脚本
./setup_go_env.sh
```

脚本将自动检测您的shell类型（bash或zsh），并在相应的配置文件中添加必要的环境变量。

## 手动配置方法

如果您想手动配置Go环境变量，请按照以下步骤操作：

### 1. 确定您的Shell类型

```bash
echo $SHELL
```

### 2. 编辑相应的配置文件

- 对于 **Zsh** (macOS Catalina及更高版本的默认shell)：

  ```bash
  nano ~/.zshrc
  ```

- 对于 **Bash**：

  ```bash
  nano ~/.bash_profile  # macOS
  # 或
  nano ~/.bashrc        # Linux
  ```

### 3. 添加以下内容到配置文件末尾

```bash
# Go环境变量配置
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### 4. 保存并退出编辑器

- 在nano编辑器中：按 `Ctrl+O` 保存，然后按 `Ctrl+X` 退出
- 在vim编辑器中：按 `Esc`，然后输入 `:wq` 并按 `Enter`

### 5. 应用更改

```bash
# 对于Zsh
source ~/.zshrc

# 对于Bash
source ~/.bash_profile  # macOS
# 或
source ~/.bashrc        # Linux
```

## 验证配置

配置完成后，您可以通过以下命令验证Go环境变量是否正确设置：

```bash
which go
go version
echo $GOPATH
```

## 开机自动加载

一旦您将环境变量添加到相应的shell配置文件（`.zshrc`、`.bash_profile`或`.bashrc`），这些设置将在每次启动新终端或重启计算机时自动加载。

## 常见问题解决

### 1. 环境变量未生效

如果配置后环境变量未生效，请尝试以下方法：

- 确保您已经保存了配置文件
- 重新加载配置文件：`source ~/.zshrc`（或相应的配置文件）
- 重启终端或开启新的终端窗口

### 2. Go命令未找到

如果配置后仍然提示 "go command not found"，请检查：

- Go是否正确安装在 `/usr/local/go` 目录
- 路径是否正确添加到配置文件
- 是否有语法错误在配置文件中

### 3. 多个Go版本

如果您有多个Go版本，可以通过修改PATH环境变量来切换使用的版本：

```bash
# 使用特定版本的Go
export PATH=/path/to/specific/go/bin:$PATH
```

## 其他配置选项

### 配置GOPROXY（国内用户推荐）

为了加速Go模块下载，国内用户可以配置GOPROXY：

```bash
# 添加到您的shell配置文件中
export GOPROXY=https://goproxy.cn,direct
```

### 配置GO111MODULE

```bash
# 添加到您的shell配置文件中
export GO111MODULE=on
```

## 结论

正确配置Go环境变量对于Go开发至关重要。通过本指南，您应该能够成功配置Go环境变量，并使其在开机时自动加载。如果您遇到任何问题，请参考Go官方文档或社区资源获取更多帮助。