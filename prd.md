兄弟，下面是基于 **Gin 技术栈** 的服务器管理项目产品方案。

# 服务器管理与健康监控系统产品方案

## 1. 产品定位

做一个轻量级服务器管理与健康监控系统。

核心能力：

```text
1. 管理多台服务器
2. 每 10 分钟通过 SSH 自动采集服务器指标
3. 采集 CPU、内存、负载、硬盘等状态
4. 支持阈值配置
5. 指标超过阈值后自动预警
6. 后台展示服务器健康状态和历史趋势
```

第一版不做复杂 Agent，不依赖 Prometheus，不接 Kubernetes，先做 **SSH 拉取式监控系统**。

---

# 2. 技术栈

## 后端

| 模块      | 技术                      |
| ------- | ----------------------- |
| HTTP 框架 | Gin                     |
| 数据库     | MySQL                   |
| ORM     | GORM                    |
| 缓存 / 队列 | Redis                   |
| SSH     | golang.org/x/crypto/ssh |
| 定时任务    | robfig/cron             |
| 日志      | zap                     |
| 配置管理    | Viper                   |
| 参数校验    | go-playground/validator |
| JWT 登录  | golang-jwt/jwt          |
| 数据库迁移   | golang-migrate          |
| API 文档  | Swagger / OpenAPI       |

## 前端

| 模块    | 技术                     |
| ----- | ---------------------- |
| 前端框架  | Vue 3                  |
| 构建工具  | Vite                   |
| UI 组件 | Naive UI / Arco Design |
| 图表    | ECharts                |
| 请求库   | Axios                  |
| 状态管理  | Pinia                  |
| 路由    | Vue Router             |

## 部署

| 模块   | 技术               |
| ---- | ---------------- |
| 后端部署 | Docker / systemd |
| 前端部署 | Nginx            |
| 数据库  | MySQL            |
| 缓存   | Redis            |
| 反向代理 | Nginx            |
| 日志收集 | 本地文件 / Loki 后续可选 |

---

# 3. 系统架构

```text
Vue3 管理后台
        |
        v
Gin API Server
        |
        | 服务器管理 / 告警规则 / 图表查询
        v
MySQL
        |
        v
Cron Scheduler
        |
        | 每 10 分钟
        v
Collector Worker Pool
        |
        | 并发 SSH
        v
目标服务器
        |
        | CPU / 内存 / 负载 / 硬盘
        v
Metric Storage
        |
        v
Alert Evaluator
        |
        v
Notifier
        |
        v
飞书 / Telegram / 邮件
```

---

# 4. 核心业务流程

## 4.1 服务器采集流程

```text
1. 定时器每 10 分钟触发
2. 查询启用中的服务器列表
3. 按服务器生成采集任务
4. Worker Pool 并发执行 SSH 采集
5. 每台服务器执行一段 Shell 脚本
6. 脚本返回 JSON
7. Go 后端解析 JSON
8. 写入 server_metrics 表
9. 判断告警规则
10. 触发通知
```

---

## 4.2 告警流程

```text
采集数据
   |
   v
读取告警规则
   |
   v
判断是否超过阈值
   |
   v
是否连续 N 次异常？
   |
   | 是
   v
创建告警事件
   |
   v
发送通知
   |
   v
告警中
   |
   v
后续采集恢复正常
   |
   v
标记已恢复
   |
   v
发送恢复通知
```

---

# 5. 功能模块设计

## 5.1 登录与权限

### 第一版功能

```text
1. 管理员登录
2. JWT 鉴权
3. 修改密码
4. 基础权限控制
```

### 后续扩展

```text
1. 多用户
2. 角色权限
3. 操作日志
4. 双因素认证
```

---

## 5.2 服务器管理

### 功能

```text
1. 新增服务器
2. 编辑服务器
3. 删除服务器
4. 启用 / 禁用服务器
5. 测试 SSH 连接
6. 手动采集一次
7. 查看服务器详情
8. 服务器分组
```

### 服务器字段

```text
服务器名称
服务器分组
公网 IP / 内网 IP
SSH 端口
SSH 用户名
认证方式
私钥
密码，非推荐
备注
是否启用
最后采集时间
当前状态
```

### 服务器状态

| 状态       | 说明       |
| -------- | -------- |
| normal   | 正常       |
| warning  | 有预警      |
| critical | 严重异常     |
| offline  | SSH 连接失败 |
| disabled | 已禁用      |
| unknown  | 未采集      |

---

## 5.3 指标采集

### 第一版采集指标

| 类型  | 指标                  |
| --- | ------------------- |
| CPU | CPU 使用率、CPU 核心数     |
| 内存  | 总内存、已用内存、可用内存、使用率   |
| 负载  | load1、load5、load15  |
| 硬盘  | 分区、挂载点、总容量、已用容量、使用率 |
| 系统  | 主机名、运行时间、系统版本       |
| 网络  | 第一版可不做，第二版增加        |

---

## 5.4 采集命令设计

不要每个指标 SSH 一次。
应该每台服务器只建立一次 SSH 连接，执行一个脚本，然后返回 JSON。

### 远程采集脚本返回结构

```json
{
  "hostname": "api-server-01",
  "os": "Ubuntu 22.04",
  "uptime_seconds": 864000,
  "cpu": {
    "usage_percent": 37.82,
    "cores": 4
  },
  "memory": {
    "total_mb": 7936,
    "used_mb": 4210,
    "available_mb": 2860,
    "usage_percent": 53.05
  },
  "load": {
    "load1": 0.82,
    "load5": 0.76,
    "load15": 0.61
  },
  "disks": [
    {
      "filesystem": "/dev/vda1",
      "mount_point": "/",
      "size_bytes": 53687091200,
      "used_bytes": 36507222016,
      "available_bytes": 17179869184,
      "usage_percent": 68
    }
  ]
}
```

---

# 6. 告警规则设计

## 6.1 告警规则类型

第一版支持：

```text
CPU 使用率
内存使用率
磁盘使用率
1 分钟负载
5 分钟负载
15 分钟负载
SSH 连接失败
```

---

## 6.2 告警条件

支持：

| 操作符 | 说明   |
| --- | ---- |
| >   | 大于   |
| >=  | 大于等于 |
| <   | 小于   |
| <=  | 小于等于 |
| ==  | 等于   |
| !=  | 不等于  |

---

## 6.3 告警等级

| 等级       | 说明 |
| -------- | -- |
| info     | 提醒 |
| warning  | 警告 |
| critical | 严重 |

---

## 6.4 默认告警规则

| 指标      |            条件 | 连续次数 | 等级       |
| ------- | ------------: | ---: | -------- |
| CPU 使用率 |         > 90% |  3 次 | warning  |
| 内存使用率   |         > 90% |  3 次 | warning  |
| 磁盘使用率   |         > 85% |  1 次 | critical |
| load5   | > CPU 核心数 * 2 |  3 次 | warning  |
| SSH 失败  |        >= 3 次 |  3 次 | critical |

说明：

```text
每 10 分钟采集一次。
连续 3 次异常 = 持续 30 分钟异常。
```

---

## 6.5 防止重复告警

不要每 10 分钟重复推送同一条告警。

建议规则：

```text
1. 第一次进入告警状态，发送通知
2. 告警持续中，不重复发送
3. 如果持续超过 1 小时，再次提醒一次
4. 恢复正常时，发送恢复通知
```

---

# 7. 页面设计

## 7.1 仪表盘

展示整体健康情况。

### 内容

```text
服务器总数
正常服务器数量
告警服务器数量
离线服务器数量
CPU 最高的服务器 Top 5
内存最高的服务器 Top 5
磁盘最高的服务器 Top 5
最近告警列表
```

---

## 7.2 服务器列表页

### 表格字段

| 字段     | 说明                     |
| ------ | ---------------------- |
| 名称     | 服务器名称                  |
| IP     | 服务器地址                  |
| 分组     | 所属分组                   |
| 状态     | 正常 / 告警 / 离线           |
| CPU    | 当前 CPU 使用率             |
| 内存     | 当前内存使用率                |
| 磁盘     | 当前最高磁盘使用率              |
| 负载     | load1 / load5 / load15 |
| 最后采集时间 | 最近一次采集时间               |
| 操作     | 详情 / 编辑 / 测试 / 采集      |

---

## 7.3 服务器详情页

### 页面内容

```text
1. 基础信息
2. 当前状态
3. CPU 趋势图
4. 内存趋势图
5. 负载趋势图
6. 硬盘使用情况
7. 最近采集记录
8. 当前告警事件
9. 操作日志
```

---

## 7.4 告警规则页

### 功能

```text
1. 新增规则
2. 编辑规则
3. 启用 / 禁用规则
4. 删除规则
5. 设置全局规则
6. 设置服务器分组规则
7. 设置单台服务器规则
```

---

## 7.5 告警事件页

### 字段

| 字段     | 说明                 |
| ------ | ------------------ |
| 服务器    | 触发服务器              |
| 指标     | CPU / 内存 / 磁盘      |
| 当前值    | 触发时数值              |
| 阈值     | 规则阈值               |
| 等级     | warning / critical |
| 状态     | 告警中 / 已恢复          |
| 首次触发时间 | 第一次异常时间            |
| 最近触发时间 | 最近一次异常时间           |
| 恢复时间   | 正常恢复时间             |

---

## 7.6 通知通道页

第一版支持：

```text
1. 飞书机器人
2. Telegram Bot
3. 邮件 SMTP
```

每个通知通道可以测试发送。

---

# 8. 数据库设计

## 8.1 servers

```sql
CREATE TABLE servers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    group_id BIGINT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 22,
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(20) NOT NULL DEFAULT 'private_key',
    private_key_encrypted TEXT NULL,
    password_encrypted TEXT NULL,
    remark VARCHAR(500) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    enabled TINYINT NOT NULL DEFAULT 1,
    last_checked_at DATETIME NULL,
    last_error TEXT NULL,
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_enabled (enabled),
    INDEX idx_status (status)
);
```

---

## 8.2 server_groups

```sql
CREATE TABLE server_groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    remark VARCHAR(500) NULL,
    created_at DATETIME,
    updated_at DATETIME
);
```

---

## 8.3 server_metrics

```sql
CREATE TABLE server_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    server_id BIGINT NOT NULL,
    hostname VARCHAR(255) NULL,
    os VARCHAR(255) NULL,
    cpu_usage DECIMAL(5,2) NULL,
    cpu_cores INT NULL,
    memory_total_mb BIGINT NULL,
    memory_used_mb BIGINT NULL,
    memory_available_mb BIGINT NULL,
    memory_usage DECIMAL(5,2) NULL,
    load_1 DECIMAL(10,2) NULL,
    load_5 DECIMAL(10,2) NULL,
    load_15 DECIMAL(10,2) NULL,
    disk_max_usage DECIMAL(5,2) NULL,
    uptime_seconds BIGINT NULL,
    raw_json JSON NULL,
    collected_at DATETIME NOT NULL,
    created_at DATETIME,
    INDEX idx_server_collected (server_id, collected_at),
    INDEX idx_collected_at (collected_at)
);
```

---

## 8.4 server_disks

```sql
CREATE TABLE server_disks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    metric_id BIGINT NOT NULL,
    server_id BIGINT NOT NULL,
    filesystem VARCHAR(255) NULL,
    mount_point VARCHAR(255) NULL,
    size_bytes BIGINT NULL,
    used_bytes BIGINT NULL,
    available_bytes BIGINT NULL,
    usage_percent DECIMAL(5,2) NULL,
    created_at DATETIME,
    INDEX idx_metric_id (metric_id),
    INDEX idx_server_id (server_id)
);
```

---

## 8.5 alert_rules

```sql
CREATE TABLE alert_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    scope_type VARCHAR(20) NOT NULL DEFAULT 'global',
    scope_id BIGINT NULL,
    metric VARCHAR(50) NOT NULL,
    operator VARCHAR(10) NOT NULL,
    threshold DECIMAL(10,2) NOT NULL,
    duration_times INT NOT NULL DEFAULT 1,
    level VARCHAR(20) NOT NULL DEFAULT 'warning',
    enabled TINYINT NOT NULL DEFAULT 1,
    notify_interval_minutes INT NOT NULL DEFAULT 60,
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_scope (scope_type, scope_id),
    INDEX idx_enabled (enabled)
);
```

### scope_type 说明

| 值      | 说明      |
| ------ | ------- |
| global | 全局规则    |
| group  | 服务器分组规则 |
| server | 单台服务器规则 |

---

## 8.6 alert_events

```sql
CREATE TABLE alert_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    server_id BIGINT NOT NULL,
    alert_rule_id BIGINT NOT NULL,
    metric VARCHAR(50) NOT NULL,
    current_value DECIMAL(10,2) NULL,
    threshold DECIMAL(10,2) NULL,
    level VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'firing',
    first_triggered_at DATETIME NOT NULL,
    last_triggered_at DATETIME NOT NULL,
    recovered_at DATETIME NULL,
    last_notified_at DATETIME NULL,
    notify_count INT NOT NULL DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_server_status (server_id, status),
    INDEX idx_rule_status (alert_rule_id, status)
);
```

---

## 8.7 notification_channels

```sql
CREATE TABLE notification_channels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSON NOT NULL,
    enabled TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME,
    updated_at DATETIME
);
```

---

## 8.8 alert_notifications

```sql
CREATE TABLE alert_notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_event_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL,
    response TEXT NULL,
    sent_at DATETIME NULL,
    created_at DATETIME,
    INDEX idx_alert_event_id (alert_event_id)
);
```

---

# 9. 后端目录结构

```text
server-manager/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── middleware/
│   │   └── handler/
│   ├── config/
│   ├── model/
│   ├── repository/
│   ├── service/
│   │   ├── collector/
│   │   ├── alert/
│   │   ├── notifier/
│   │   └── auth/
│   ├── scheduler/
│   ├── worker/
│   ├── sshclient/
│   ├── response/
│   └── utils/
├── migrations/
├── scripts/
├── configs/
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

---

# 10. 核心服务划分

## 10.1 CollectorService

负责采集服务器指标。

```text
1. 获取服务器 SSH 信息
2. 建立 SSH 连接
3. 执行采集脚本
4. 解析 JSON
5. 写入 metrics
6. 更新服务器状态
```

---

## 10.2 AlertEvaluator

负责判断是否触发告警。

```text
1. 查询服务器适用的告警规则
2. 判断当前值是否超过阈值
3. 判断是否连续 N 次异常
4. 创建或更新 alert_events
5. 处理恢复状态
```

---

## 10.3 Notifier

负责发送通知。

```text
1. 飞书机器人
2. Telegram Bot
3. 邮件 SMTP
4. 记录通知结果
```

---

## 10.4 Scheduler

负责定时触发采集。

```text
1. 每 10 分钟执行一次全量采集
2. 支持手动触发单台服务器采集
3. 控制采集并发数量
4. 防止任务重叠
```

---

# 11. 并发采集设计

假设有 100 台服务器，不要串行采集。

建议使用 Worker Pool。

```text
1. 定时器触发
2. 查询所有启用服务器
3. 写入任务队列 channel
4. 启动固定数量 worker
5. 每个 worker 处理一台服务器
6. 全部完成后记录采集批次结果
```

### 推荐配置

```yaml
collector:
  interval: "10m"
  worker_count: 20
  ssh_connect_timeout: "5s"
  ssh_command_timeout: "15s"
  max_retries: 1
```

### 规模建议

|       服务器数量 |  worker_count |
| ----------: | ------------: |
|      10 台以内 |             5 |
|  10 - 100 台 |            20 |
| 100 - 500 台 |            50 |
|     500 台以上 | 拆分多 Collector |

---

# 12. 配置文件设计

```yaml
app:
  name: server-manager
  env: production
  port: 8080

database:
  driver: mysql
  dsn: "user:password@tcp(127.0.0.1:3306)/server_manager?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0

jwt:
  secret: "change-me"
  expire_hours: 24

collector:
  interval: "10m"
  worker_count: 20
  ssh_connect_timeout: "5s"
  ssh_command_timeout: "15s"
  max_retries: 1

alert:
  default_notify_interval_minutes: 60

security:
  encryption_key: "32-byte-key-here"
```

---

# 13. API 设计

## 13.1 登录

```text
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/profile
```

---

## 13.2 仪表盘

```text
GET /api/dashboard/summary
GET /api/dashboard/top-cpu
GET /api/dashboard/top-memory
GET /api/dashboard/top-disk
GET /api/dashboard/recent-alerts
```

---

## 13.3 服务器

```text
GET    /api/servers
POST   /api/servers
GET    /api/servers/:id
PUT    /api/servers/:id
DELETE /api/servers/:id

POST   /api/servers/:id/test-ssh
POST   /api/servers/:id/collect
GET    /api/servers/:id/metrics
GET    /api/servers/:id/disks
GET    /api/servers/:id/alerts
```

---

## 13.4 分组

```text
GET    /api/server-groups
POST   /api/server-groups
PUT    /api/server-groups/:id
DELETE /api/server-groups/:id
```

---

## 13.5 告警规则

```text
GET    /api/alert-rules
POST   /api/alert-rules
PUT    /api/alert-rules/:id
DELETE /api/alert-rules/:id
POST   /api/alert-rules/:id/enable
POST   /api/alert-rules/:id/disable
```

---

## 13.6 告警事件

```text
GET  /api/alert-events
GET  /api/alert-events/:id
POST /api/alert-events/:id/ack
POST /api/alert-events/:id/close
```

---

## 13.7 通知通道

```text
GET    /api/notification-channels
POST   /api/notification-channels
PUT    /api/notification-channels/:id
DELETE /api/notification-channels/:id
POST   /api/notification-channels/:id/test
```

---

# 14. 后台菜单设计

```text
仪表盘
服务器管理
  - 服务器列表
  - 服务器分组
监控数据
  - 指标趋势
  - 磁盘详情
告警中心
  - 告警规则
  - 告警事件
  - 通知记录
通知通道
系统设置
  - 管理员账号
  - 系统配置
```

---

# 15. 安全设计

## 15.1 SSH 密钥安全

重点：

```text
1. 私钥必须加密后入库
2. 不建议保存 SSH 密码
3. 支持上传私钥
4. 支持私钥 passphrase
5. 私钥不可在前端明文回显
```

服务器编辑页：

```text
新增时可以填写私钥
编辑时只显示：已配置私钥
如需修改，需要重新输入
```

---

## 15.2 目标服务器权限

建议在目标服务器创建低权限用户：

```bash
sudo useradd -m monitor
```

只允许执行读取命令：

```text
free
df
cat /proc/loadavg
cat /proc/stat
uname
hostname
```

第一版不做自动修复，不要给 sudo 权限。

---

## 15.3 后台安全

```text
1. JWT 登录
2. 密码 bcrypt 加密
3. 接口权限中间件
4. 操作日志
5. 登录失败限制
6. 敏感配置加密
```

---

# 16. 告警通知文案

## 16.1 告警通知

```text
【服务器告警】

服务器：api-server-01
IP：1.2.3.4
指标：CPU 使用率
当前值：94.6%
阈值：90%
级别：warning
持续：30 分钟
时间：2026-06-15 16:30:00
```

---

## 16.2 恢复通知

```text
【服务器恢复】

服务器：api-server-01
IP：1.2.3.4
指标：CPU 使用率
当前值：42.1%
恢复时间：2026-06-15 17:10:00
```

---

## 16.3 SSH 失败通知

```text
【服务器离线】

服务器：api-server-01
IP：1.2.3.4
原因：SSH 连接失败
连续失败：3 次
时间：2026-06-15 16:30:00
```

---

# 17. MVP 版本范围

## 第一版必须做

```text
1. 管理员登录
2. 服务器管理
3. SSH 连接测试
4. 每 10 分钟自动采集
5. 手动采集
6. CPU / 内存 / 负载 / 硬盘采集
7. 最新状态展示
8. 历史趋势图
9. 告警规则
10. 飞书机器人通知
11. 告警事件列表
```

---

## 第一版暂不做

```text
1. Agent 模式
2. 自动修复
3. Kubernetes 监控
4. Prometheus 兼容
5. 多租户
6. 复杂权限系统
7. 手机 App
8. 网络流量统计
9. 进程级监控
10. 日志监控
```

---

# 18. 后续版本规划

## V1：SSH 监控版

```text
目标：快速可用
核心：服务器管理、SSH 采集、阈值告警、飞书通知
```

---

## V2：增强监控版

```text
1. 支持网络流量
2. 支持进程 Top N
3. 支持服务存活检测
4. 支持端口检测
5. 支持 HTTP 接口健康检查
6. 支持告警静默
7. 支持告警确认
```

---

## V3：Agent 版

```text
1. 每台服务器安装 Agent
2. Agent 主动上报数据
3. 支持更高频率采集
4. 支持断网缓存
5. 支持日志采集
6. 支持自动更新 Agent
```

---

## V4：产品化版本

```text
1. 多用户
2. 多团队
3. 权限系统
4. 审计日志
5. 多区域 Collector
6. Prometheus Exporter
7. Grafana 对接
8. 商业化套餐
```

---

# 19. 开发优先级

## 第一阶段：基础框架

```text
1. Gin 项目初始化
2. 配置文件
3. MySQL 连接
4. GORM 模型
5. JWT 登录
6. 基础路由
```

---

## 第二阶段：服务器管理

```text
1. servers 表
2. server_groups 表
3. 新增 / 编辑 / 删除服务器
4. SSH 信息加密保存
5. SSH 连接测试
```

---

## 第三阶段：采集系统

```text
1. SSH Client
2. 采集脚本
3. JSON 解析
4. server_metrics 入库
5. server_disks 入库
6. 手动采集接口
7. 定时采集任务
```

---

## 第四阶段：告警系统

```text
1. alert_rules 表
2. alert_events 表
3. 阈值判断
4. 连续触发判断
5. 告警恢复判断
6. 防重复通知
```

---

## 第五阶段：通知系统

```text
1. notification_channels 表
2. 飞书机器人
3. 测试通知
4. 告警通知
5. 恢复通知
```

---

## 第六阶段：前端后台

```text
1. 登录页
2. 仪表盘
3. 服务器列表
4. 服务器详情
5. 指标趋势图
6. 告警规则页
7. 告警事件页
8. 通知通道页
```

---

# 20. 项目最终形态

推荐最终定义为：

```text
一个基于 Go + Gin 的轻量服务器管理与健康监控系统。

系统通过 SSH 定时连接用户指定的服务器，采集 CPU、内存、系统负载、硬盘等指标，并将数据存储到 MySQL。用户可以在后台查看服务器实时状态和历史趋势，并配置阈值规则。当指标超过阈值并持续达到指定次数后，系统会通过飞书、Telegram 或邮件发起预警，同时在恢复正常时发送恢复通知。
```

---

# 21. 最终技术选型结论

建议第一版就这样定：

```text
后端：Go + Gin
数据库：MySQL
ORM：GORM
缓存：Redis
定时任务：robfig/cron
SSH：golang.org/x/crypto/ssh
前端：Vue3 + Vite + Naive UI
图表：ECharts
通知：飞书机器人优先
部署：Docker Compose
```

这个方案比 PHP 更适合你的场景，尤其是：

```text
1. 多服务器并发 SSH
2. 长时间运行任务
3. 定时采集
4. 后续拆分 Agent
5. 单二进制部署
```
