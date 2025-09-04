# Go 漏洞代码环境构建 Makefile

# 所有环境列表（动态从含 Dockerfile 的子目录读取）
# 取形如 bad_totp/Dockerfile -> bad_totp
ENVIRONMENTS := $(patsubst %/Dockerfile,%,$(wildcard */Dockerfile))

# 镜像版本
VERSION = v0.0.1
REGISTRY = cyberpoc

# 默认目标
.PHONY: help
help:
	@echo "可用命令："
	@echo "  make build ENV=<环境名>     - 构建指定环境的镜像"
	@echo "  make push ENV=<环境名>      - 构建并推送指定环境的镜像"
	@echo "  make build-all             - 构建所有环境的镜像"
	@echo "  make push-all              - 构建并推送所有环境的镜像"
	@echo "  make clean                 - 清理所有镜像"
	@echo "  make list                  - 列出所有环境"
	@echo ""
	@echo "可用环境："
	@echo "  $(ENVIRONMENTS)"

# 构建单个环境
.PHONY: build
build:
	@if [ -z "$(ENV)" ]; then \
		echo "错误: 请指定环境名，例如 make build ENV=exec"; \
		exit 1; \
	fi
	@if [ ! -d "$(ENV)" ]; then \
		echo "错误: 环境 $(ENV) 不存在"; \
		exit 1; \
	fi
	@echo "构建环境: $(ENV)"
	@cd $(ENV) && ../build-image-common.sh $(REGISTRY)/$(ENV):$(VERSION)

# 构建并推送单个环境
.PHONY: push
push:
	@if [ -z "$(ENV)" ]; then \
		echo "错误: 请指定环境名，例如 make push ENV=exec"; \
		exit 1; \
	fi
	@if [ ! -d "$(ENV)" ]; then \
		echo "错误: 环境 $(ENV) 不存在"; \
		exit 1; \
	fi
	@echo "构建并推送环境: $(ENV)"
	@cd $(ENV) && ../build-image-common.sh $(REGISTRY)/$(ENV):$(VERSION) push

# 构建所有环境
.PHONY: build-all
build-all:
	@echo "开始构建所有环境..."
	@for env in $(ENVIRONMENTS); do \
		if [ -d "$$env" ]; then \
			echo "构建环境: $$env"; \
			cd $$env && ../build-image-common.sh $(REGISTRY)/$$env:$(VERSION) && cd ..; \
		else \
			echo "跳过不存在的环境: $$env"; \
		fi; \
	done
	@echo "所有环境构建完成!"

# 构建并推送所有环境
.PHONY: push-all
push-all:
	@echo "开始构建并推送所有环境..."
	@for env in $(ENVIRONMENTS); do \
		if [ -d "$$env" ]; then \
			echo "构建并推送环境: $$env"; \
			cd $$env && ../build-image-common.sh $(REGISTRY)/$$env:$(VERSION) push && cd ..; \
		else \
			echo "跳过不存在的环境: $$env"; \
		fi; \
	done
	@echo "所有环境构建并推送完成!"

# 清理镜像
.PHONY: clean
clean:
	@echo "清理所有构建的镜像..."
	@for env in $(ENVIRONMENTS); do \
		docker rmi $(REGISTRY)/$$env:$(VERSION) 2>/dev/null || true; \
	done
	@echo "镜像清理完成!"

# 列出所有环境
.PHONY: list
list:
	@echo "可用环境列表："
	@for env in $(ENVIRONMENTS); do \
		if [ -d "$$env" ]; then \
			echo "  ✓ $$env"; \
		else \
			echo "  ✗ $$env (不存在)"; \
		fi; \
	done

# 运行单个环境 (仅本地测试)
.PHONY: run
run:
	@if [ -z "$(ENV)" ]; then \
		echo "错误: 请指定环境名，例如 make run ENV=exec PORT=8080"; \
		exit 1; \
	fi
	@if [ -z "$(PORT)" ]; then \
		echo "错误: 请指定端口，例如 make run ENV=exec PORT=8080"; \
		exit 1; \
	fi
	@echo "运行环境 $(ENV) 在端口 $(PORT)..."
	@docker run --rm -p $(PORT):80 -e flag="CTF{test_flag}" $(REGISTRY)/$(ENV):$(VERSION)

# 查看环境状态
.PHONY: status
status:
	@echo "Docker 镜像状态："
	@docker images | grep "$(REGISTRY)" | grep "$(VERSION)" || echo "未找到相关镜像"
