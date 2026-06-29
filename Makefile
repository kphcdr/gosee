.PHONY: all frontend backend run dev typecheck vet test clean tidy deploy publish help

# ===== 生产发布配置（命令行可覆盖，例如 make publish SSH_TARGET=root@host）=====
SSH_TARGET ?= root@114.215.131.37
REMOTE_DIR ?= /data/www/system.kphcdr.com
SERVICE_NAME ?= gosee
HEALTH_URL ?= http://127.0.0.1:8910/health

# ===== 构建 =====

# 构建单二进制（前端 embed 进后端）
all: frontend backend

# 构建前端（输出 web/dist）
frontend:
	cd web && pnpm install && pnpm build

# 构建后端二进制（自动 embed web/dist）
backend:
	go build -o gosee ./cmd/server

# 交叉编译 Linux amd64 单二进制（在 mac 上构建，部署到 Linux 服务器）
# glebarez/sqlite 纯 Go 免 CGO，无需 C 工具链
build-linux: frontend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gosee-linux-amd64 ./cmd/server

# ===== 开发 =====

# 运行后端（go run，读 configs/config.yaml）
run:
	go run ./cmd/server

# 前端开发模式（Vite dev server :5173，需另开终端跑 make run）
dev:
	cd web && pnpm dev

# ===== 检查 =====

# 前端类型检查
typecheck:
	cd web && pnpm exec vue-tsc -b

# 后端静态检查
vet:
	go vet ./...

# 全量检查
check: typecheck vet
	@echo "✅ 前后端检查通过"

# ===== 维护 =====

# 清理构建产物
clean:
	rm -f gosee
	rm -rf web/dist

# Go 依赖整理
tidy:
	go mod tidy

# 前端依赖更新
install:
	cd web && pnpm install

# ===== 部署产物（构建到 deploy/dist）=====
deploy: all
	@echo "✅ 部署产物：gosee + configs/config.prod.yaml + deploy/"

# 一条龙发布到生产：检查 → 构建 → 上传 → 备份替换 → 重启 → 健康检查
# 仅更新二进制，不覆盖线上 config.yaml、gosee.db 和日志。
publish: check build-linux
	scp gosee-linux-amd64 $(SSH_TARGET):$(REMOTE_DIR)/.$(SERVICE_NAME).new
	ssh $(SSH_TARGET) 'set -e; \
		cd $(REMOTE_DIR); \
		chmod +x .$(SERVICE_NAME).new; \
		if [ -f $(SERVICE_NAME) ]; then cp -p $(SERVICE_NAME) $(SERVICE_NAME).bak; fi; \
		mv .$(SERVICE_NAME).new $(SERVICE_NAME); \
		systemctl restart $(SERVICE_NAME); \
		for i in $$(seq 1 15); do \
			if curl -fsS $(HEALTH_URL) >/dev/null; then \
				echo "✅ $(SERVICE_NAME) 发布成功，健康检查通过"; \
				systemctl status $(SERVICE_NAME) --no-pager; \
				exit 0; \
			fi; \
			sleep 1; \
		done; \
		echo "❌ 健康检查失败，正在回滚"; \
		if [ -f $(SERVICE_NAME).bak ]; then \
			mv $(SERVICE_NAME).bak $(SERVICE_NAME); \
			systemctl restart $(SERVICE_NAME); \
		fi; \
		systemctl status $(SERVICE_NAME) --no-pager || true; \
		exit 1'

# ===== 帮助 =====
help:
	@echo "gosee Makefile —— 常用命令"
	@echo ""
	@echo "  make all        构建单二进制（前端 + 后端，推荐部署用）"
	@echo "  make frontend   仅构建前端 → web/dist"
	@echo "  make backend    仅构建后端二进制（embed 前端）"
	@echo "  make run        运行后端（go run，开发）"
	@echo "  make dev        前端开发模式（Vite :5173）"
	@echo "  make check      前后端全量检查（typecheck + vet）"
	@echo "  make clean      清理构建产物"
	@echo "  make tidy       go mod tidy"
	@echo "  make build-linux 交叉编译 Linux amd64 单二进制（部署到 Linux 服务器）"
	@echo "  make deploy     构建部署产物"
	@echo "  make publish    一条龙发布到生产并重启、检查 gosee 服务"
