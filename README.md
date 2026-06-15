# gosee — 轻量级服务器监控告警系统

基于 **Go + Gin** 的 SSH 拉取式服务器监控告警系统。通过 SSH 定时采集 CPU / 内存 / 负载 / 硬盘等指标，超阈值自动触发告警并通过飞书通知。

> 完整需求见 [prd.md](./prd.md) · 部署见 [DEPLOY.md](./DEPLOY.md) · 任务跟踪见 [todo.md](./todo.md)

---

## 功能概览

**采集 → 存储 → 告警 → 通知 → 仪表盘**，完整闭环（PRD 阶段 1–8 + 部署，全部完成并验证）：

- **指标采集**：SSH 拉取 CPU/内存/负载/磁盘/系统信息，单连接脚本输出 JSON
- **定时调度**：robfig/cron 定时 + Worker Pool 并发采集，防重叠
- **告警系统**：规则（global/group/server 三级 scope）+ 阈值判断 + 连续 N 次触发 + 采集后自动评估 + 事件 ack/close
- **通知系统**：飞书 webhook（含可选签名校验），告警/恢复/离线文案，notify_interval 防重复
- **仪表盘**：服务器状态汇总 + CPU/内存/磁盘 Top5 + 最近告警
- **前端后台**：Vue3 + TS + Naive UI + ECharts，10 个页面，全真实接口
- **部署**：Go embed 前端，单二进制 + systemd + Nginx

---

## 技术栈

- **后端**：Go + Gin + GORM + Viper + zap + robfig/cron + golang.org/x/crypto/ssh
- **数据库**：GORM（`glebarez/sqlite` 免 CGO / MySQL 可切换）
- **前端**：Vue3 + TypeScript + Vite + Naive UI + ECharts + Pinia + Vue Router + unplugin 自动导入
- **安全**：AES-256-GCM（SSH 凭证）+ bcrypt（登录密码）+ JWT

---

## 快速开始

```bash
make all            # 一键构建单二进制（前端 embed 进后端）→ gosee
./gosee             # 运行（读 configs/config.yaml），监听 :8080
```

> 手动等价：`cd web && pnpm install && pnpm build` → `go build -o gosee ./cmd/server`

启动后：
- 浏览器访问 `http://localhost:8080`（前端 + API 同源）
- 自动建库（`gosee.db`）+ 迁移表结构 + 默认告警规则
- 首次启动创建管理员 `admin / admin123`

### 开发模式

```bash
make run            # 后端 go run :8080
make dev            # 前端 Vite :5173（另开终端，自动代理 /api → :8080）
make check          # 前后端检查（vue-tsc + go vet）
```

### 切换 MySQL

改 `configs/config.yaml`：
```yaml
database:
  driver: mysql
  dsn: "root:pass@tcp(127.0.0.1:3306)/gosee?charset=utf8mb4&parseTime=True&loc=Local"
```
重启即自动迁移，代码无需改动。

---

## API 接口

前缀 `/api`，除登录外均需 `Authorization: Bearer <token>`。统一响应 `{ code, message, data }`，`code=0` 成功。

**认证**
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/auth/login` | 登录，返回 JWT |
| GET | `/api/auth/profile` | 当前用户信息 |
| PUT | `/api/auth/password` | 修改密码 |

**服务器 & 分组**
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/api/server-groups` | 分组列表/新增 |
| PUT/DELETE | `/api/server-groups/:id` | 编辑/删除（有服务器时拒绝） |
| GET | `/api/servers` | 服务器列表（page/page_size/group_id/enabled/keyword） |
| POST | `/api/servers` | 新增服务器 |
| GET/PUT/DELETE | `/api/servers/:id` | 详情/编辑/删除 |
| POST | `/api/servers/:id/test-ssh` | 测试 SSH 连接 |
| POST | `/api/servers/:id/collect` | 手动采集一次 |
| GET | `/api/servers/:id/metrics` | 指标趋势（hours/limit） |
| GET | `/api/servers/:id/disks` | 最新磁盘明细 |
| POST | `/api/collect/run` | 手动全量采集（返回批次统计） |

**告警**
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/api/alert-rules` | 规则列表/新增 |
| PUT/DELETE | `/api/alert-rules/:id` | 编辑/删除 |
| POST | `/api/alert-rules/:id/enable` `/disable` | 启用/禁用 |
| GET | `/api/alert-events` | 事件列表 |
| POST | `/api/alert-events/:id/ack` `/close` | 确认/关闭 |

**通知**
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/api/notification-channels` | 通道列表/新增 |
| PUT/DELETE | `/api/notification-channels/:id` | 编辑/删除 |
| POST | `/api/notification-channels/:id/test` | 测试发送（飞书） |

**仪表盘**
| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/dashboard/summary` | 状态汇总（总数/正常/告警/离线） |
| GET | `/api/dashboard/top-cpu` `/top-memory` `/top-disk` | 各指标 Top5 |
| GET | `/api/dashboard/recent-alerts` | 最近告警 |

---

## 目录结构

```
cmd/server/main.go              # 入口：装配依赖 + 优雅启停
internal/
├── api/
│   ├── router.go               # 路由 + 前端 SPA 托管（NoRoute 兜底）
│   ├── handler/                # auth/server/server_group/collector/alert_*/notification_channel/dashboard
│   └── middleware/             # jwt / cors / recovery
├── config/                     # Viper 配置
├── database/                   # GORM + AutoMigrate + 种子（管理员/默认告警规则）
├── model/                      # user/server/server_group/server_metric/server_disk/alert_rule/alert_event/notification_channel/alert_notification
├── repository/                 # 数据访问层
├── scheduler/                  # 定时采集（cron + Worker Pool + 防重叠）
├── service/
│   ├── auth/                   # 登录/JWT
│   ├── collector/              # 指标采集 + AlertHook（采集后触发评估）
│   ├── server/ server_group/   # 服务器/分组
│   ├── alert/                  # 告警规则/事件/评估器
│   ├── notifier/               # 飞书通知
│   └── dashboard/              # 仪表盘聚合
├── sshclient/                  # SSH 连接与命令执行
├── response/                   # 统一响应封装
└── utils/                      # crypto / hash / jwt / logger
web/                            # 前端（详见 web/README.md）
├── embed.go                    # //go:embed dist（前端打进二进制）
└── src/                        # Vue3 源码
configs/                        # config.yaml(开发) / config.prod.yaml(生产模板)
deploy/                         # gosee.service(systemd) / nginx.conf
Makefile / DEPLOY.md / todo.md
```

---

## 安全说明

- SSH 私钥/密码 **AES-256-GCM** 加密入库，密钥 `security.encryption_key`（**生产必须替换**，否则等于明文存储）
- 凭证字段 `json:"-"`，接口不回显；编辑时空值 = 保留旧值
- 登录密码 bcrypt 存储；JWT HS256
- ⚠️ SSH 连接当前 `InsecureIgnoreHostKey`（未校验主机指纹，见 todo 技术债）

---

## 部署

单二进制（Go embed 前端 + SQLite + systemd）。完整步骤见 **[DEPLOY.md](./DEPLOY.md)**。

```bash
make all    # 构建单二进制 gosee（含前端）
```

**生产必做**：替换 `jwt.secret` 与 `security.encryption_key`（`openssl rand -hex 32`）。

---

## 已知限制 / 后续

见 [todo.md](./todo.md) 技术债：SSH host key 校验、ssh_fail 连续计数、Telegram/邮件通知、指标数据保留、登录限流、操作日志等。
