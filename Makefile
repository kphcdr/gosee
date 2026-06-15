.PHONY: all frontend backend run dev typecheck vet test clean tidy deploy help

# ===== 构建 =====

# 构建单二进制（前端 embed 进后端）
all: frontend backend

# 构建前端（输出 web/dist）
frontend:
	cd web && pnpm install && pnpm build

# 构建后端二进制（自动 embed web/dist）
backend:
	go build -o gosee ./cmd/server

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
	@echo "  make deploy     构建部署产物"
