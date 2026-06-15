# gosee — 轻量级服务器管理与健康监控系统

基于 **Go + Gin** 的 SSH 拉取式服务器监控告警系统。通过 SSH 定时采集 CPU / 内存 / 负载 / 硬盘等指标，存储到数据库，支持阈值规则与多通道告警。

> 完整需求见 [prd.md](./prd.md)

---

## 当前进度

✅ **第一阶段：基础框架** — 已完成并验证
✅ **第二阶段：服务器管理** — 已完成并验证
✅ **第三阶段：指标采集** — 已完成并验证（真实 SSH 端到端跑通）
✅ **第四阶段：定时调度** — 已完成并验证（cron + Worker Pool 并发）
✅ **第八阶段：前端后台** — 已完成（Vue3 + TS + Naive UI + ECharts，10 个页面，dashboard/告警/通知用前端 mock 待后端对接）
✅ **第五阶段：告警系统** — 已完成（规则/事件/评估器，采集后自动评估，端到端验证通过）
✅ **第六阶段：通知系统（飞书）** — 已完成（通道 CRUD + 飞书 webhook + 文案 + notify_interval 防重复 + 告警自动通知）
✅ **第七阶段：仪表盘** — 已完成（summary + CPU/内存/磁盘 Top5 + 最近告警，真实聚合）

| 能力 | 状态 |
| --- | --- |
| Gin 项目骨架 / 配置(Viper) / 日志(zap) | ✅ |
| 数据库（GORM，SQLite/MySQL 可切）+ 自动迁移 | ✅ |
| 管理员登录（JWT）+ 修改密码 | ✅ |
| 统一响应 / Recovery / CORS 中间件 | ✅ |
| 服务器分组 CRUD | ✅ |
| 服务器 CRUD | ✅ |
| SSH 私钥/密码 AES-256-GCM 加密入库，前端不回显明文 | ✅ |
| SSH 连接测试（成功/失败均更新状态与错误） | ✅ |
| SSH 指标采集（CPU/内存/负载/磁盘/系统）+ 手动采集接口 | ✅ |
| 指标趋势查询 + 磁盘明细 | ✅ |
| 定时自动采集（robfig/cron，间隔可配）+ 任务防重叠 | ✅ |
| Worker Pool 并发采集（worker 数可配） | ✅ |
| 前端后台 Vue3 + TS（登录/布局/服务器/分组/详情趋势图） | ✅ |
| 仪表盘/告警规则/事件/通知通道（前端 mock，待后端对接） | ✅ |
| 告警系统（规则 CRUD + 评估器 + 采集后自动评估 + 事件 ack/close） | ✅ |
| 通知系统（飞书 webhook + 告警/恢复/离线文案 + notify_interval 防重复） | ✅ |
| 仪表盘（服务器状态汇总 + CPU/内存/磁盘 Top5 + 最近告警，真实聚合） | ✅ |

⏳ **部署（待做）**：Docker / Go embed 前端 / Nginx / 生产加固

---

## 技术栈

- **HTTP 框架**：Gin
- **ORM**：GORM（驱动可配：`sqlite` 本地开发 / `mysql` 生产）
- **SQLite 驱动**：`glebarez/sqlite`（纯 Go，**免 CGO**）
- **配置**：Viper（支持环境变量覆盖，前缀 `GOSEE_`）
- **日志**：zap（控制台 + 文件双输出）
- **认证**：golang-jwt/jwt v5 + bcrypt
- **SSH**：golang.org/x/crypto/ssh
- **加密**：AES-256-GCM（私钥、密码落库加密）

---

## 快速开始

```bash
# 拉依赖 + 编译
go mod tidy
go build -o bin/gosee ./cmd/server

# 启动（默认读 configs/config.yaml）
./bin/gosee
# 或指定配置：GOSEE_CONFIG=configs/config.prod.yaml ./bin/gosee
```

启动后：
- 监听 `:8080`
- 自动建库（`gosee.db`）+ 迁移表结构
- 首次启动自动创建管理员 `admin / admin123`（见 `configs/config.yaml`）

健康检查：`GET http://localhost:8080/health`

### 切换到 MySQL

改 `configs/config.yaml`：
```yaml
database:
  driver: mysql
  dsn: "root:pass@tcp(127.0.0.1:3306)/gosee?charset=utf8mb4&parseTime=True&loc=Local"
```
代码无需改动。

---

## API（已实现）

所有接口前缀 `/api`，除登录外均需 `Authorization: Bearer <token>`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/auth/login` | 登录，返回 JWT |
| GET | `/api/auth/profile` | 当前用户信息 |
| PUT | `/api/auth/password` | 修改密码 |
| GET | `/api/server-groups` | 分组列表 |
| POST | `/api/server-groups` | 新增分组 |
| PUT | `/api/server-groups/:id` | 编辑分组 |
| DELETE | `/api/server-groups/:id` | 删除分组（有服务器时拒绝） |
| GET | `/api/servers` | 服务器列表（分页/分组/关键字过滤） |
| POST | `/api/servers` | 新增服务器 |
| GET | `/api/servers/:id` | 服务器详情 |
| PUT | `/api/servers/:id` | 编辑服务器 |
| DELETE | `/api/servers/:id` | 删除服务器 |
| POST | `/api/servers/:id/test-ssh` | 测试 SSH 连接 |
| POST | `/api/servers/:id/collect` | 手动采集一次指标 |
| GET | `/api/servers/:id/metrics` | 指标趋势（默认最近 24h） |
| GET | `/api/servers/:id/disks` | 最新磁盘明细 |
| POST | `/api/collect/run` | 手动触发全量采集（返回批次统计） |

统一响应：`{ "code": 0, "message": "success", "data": ... }`，`code=0` 成功。

---

## 目录结构

```
cmd/server/main.go            # 入口：装配依赖 + 优雅启停
internal/
├── api/
│   ├── router.go             # 路由注册
│   ├── handler/              # HTTP handler（auth/server/server_group）
│   └── middleware/           # jwt / cors / recovery
├── config/                   # Viper 配置
├── database/                 # GORM 初始化 + AutoMigrate + 默认管理员
├── model/                    # GORM 模型（user/server/server_group/server_metric）
├── repository/               # 数据访问层
├── scheduler/                # 定时采集调度（cron + Worker Pool）
├── service/                  # 业务层
│   ├── auth/                 # 登录/JWT
│   ├── collector/            # 指标采集（脚本+解析+入库）
│   ├── server/               # 服务器管理
│   └── server_group/         # 分组
├── sshclient/                # SSH 连接与命令执行
├── response/                 # 统一响应封装
└── utils/                    # crypto(AES-GCM) / hash(bcrypt) / jwt / logger
configs/config.yaml           # 配置文件
```

---

## 安全说明

- SSH 私钥/密码使用 **AES-256-GCM** 加密后入库，密钥配置在 `security.encryption_key`（**生产环境必须替换**）。
- 接口不回显私钥/密码明文（模型字段 `json:"-"`）。
- 编辑服务器时，未提供新凭证则保留旧值。
- 密码使用 **bcrypt** 存储。

---

## 前端后台

前端在 `web/` 目录（Vue3 + TypeScript + Vite + Naive UI + ECharts + Pinia + Vue Router）。

```bash
cd web
pnpm install
pnpm dev      # http://localhost:5173，自动代理 /api → 后端 :8080
pnpm build    # 生产构建，输出 web/dist/
```

开发时同时启动后端（`go run ./cmd/server`），登录 `admin / admin123`。

### 数据对接

所有页面均已对接真实后端接口（仪表盘 / 告警规则 / 告警事件 / 通知通道在阶段五~七实现）。`web/src/api/mock/` 目录与 `VITE_ENABLE_MOCK` 开关已无引用，可安全删除（保留不影响运行）。

### 已知偏差

- **服务器列表不显示实时 CPU/内存/磁盘列**：后端 `/api/servers` 的 `Server` 模型不含实时指标字段。实时指标在**服务器详情页**展示（趋势图 + 当前值）。

---

## 部署

采用 **单二进制部署**：Go embed 把前端打进后端二进制（~40M），SQLite 文件存储，systemd 守护。完整步骤见 **[DEPLOY.md](./DEPLOY.md)**。

```bash
# 1. 构建前端
cd web && pnpm install && pnpm build        # → web/dist
# 2. 构建后端（自动 embed 前端）
cd .. && go build -o gosee ./cmd/server      # → 单二进制 gosee
```

部署产物：`gosee`（二进制）+ `configs/config.prod.yaml`（配置）+ `deploy/`（systemd / nginx）。Nginx 只需反代到 8080（前端 + API 同源）。

**生产必做**：替换 `jwt.secret` 与 `security.encryption_key`（`openssl rand -hex 32`），见 DEPLOY.md 加固清单。
