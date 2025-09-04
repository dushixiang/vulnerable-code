# GitHub Actions 自动化构建

本项目使用 GitHub Actions 自动构建和推送 Docker 镜像到 Docker Hub。

## 🔧 设置说明

### 1. Docker Hub Secrets 配置

在 GitHub 仓库的 Settings > Secrets and variables > Actions 中添加以下 secrets：

- `DOCKER_USERNAME`: Docker Hub 用户名
- `DOCKER_PASSWORD`: Docker Hub 访问令牌 (推荐) 或密码

### 2. Docker Hub 访问令牌生成

1. 登录 [Docker Hub](https://hub.docker.com)
2. 进入 Account Settings > Security
3. 点击 "New Access Token"
4. 选择权限 (建议选择 "Read, Write, Delete")
5. 复制生成的令牌作为 `DOCKER_PASSWORD`

## 📋 工作流说明

### docker-build-push.yml

**触发条件:**
- 推送到 `main`/`master` 分支且 `go/` 目录有变更
- Pull Request 涉及 `go/` 目录
- 手动触发 (可选择特定环境)

**功能:**
- 自动检测所有包含 Dockerfile 的环境
- 并行构建多架构镜像 (linux/amd64, linux/arm64)
- 推送到 Docker Hub (仅非 PR 时)
- 安全扫描 (使用 Trivy)
- 生成构建摘要

**镜像标签:**
- `latest` (主分支)
- `v0.0.1` (主分支)
- `<branch>-<sha>` (所有分支)
- `pr-<number>` (Pull Request)

### cleanup-images.yml

**触发条件:**
- 每周日凌晨 2 点自动运行
- 手动触发

**功能:**
- 清理旧的 Docker 镜像标签
- 保留每个环境最新的 5 个版本

## 🚀 使用方式

### 自动触发
提交代码到主分支即可自动触发构建。

### 手动触发
1. 进入 GitHub 仓库的 Actions 页面
2. 选择 "Build and Push Docker Images" 工作流
3. 点击 "Run workflow"
4. 可选择构建特定环境或全部环境

### 构建特定环境
```bash
# 在 workflow_dispatch 输入框中指定环境名
environment: exec
```
