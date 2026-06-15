# gosee 开发任务清单

> 基于 [prd.md](./prd.md)，按阶段拆分。完成进度：**阶段 1–8 全部交付并验证**（核心功能闭环完成）。

---

## ✅ 已完成

- [x] **阶段一：基础框架** — Gin 骨架 / Viper 配置 / zap 日志 / GORM(SQLite/MySQL 可切) / JWT 登录 / 统一响应 / 中间件
- [x] **阶段二：服务器管理** — 服务器 & 分组 CRUD / 私钥密码 AES-GCM 加密入库 / SSH 连接测试
- [x] **阶段三：指标采集** — 采集脚本(单连接输出 JSON) / metrics+disks 入库 / 手动采集 / 趋势查询 / 磁盘明细
- [x] **阶段四：定时调度** — robfig/cron 定时 / Worker Pool 并发 / 防重叠 / SSH 命令超时 / 手动全量采集
- [x] **阶段五：告警系统** — 模型/规则/事件/评估器/采集后自动评估，端到端验证通过
  - alert_rules / alert_events 模型 + 迁移；默认规则种子；规则 CRUD/启停 + 事件查询/ack/close
  - AlertEvaluator：scope 三级 + 阈值(6 运算符) + 连续 N 次 + 触发/恢复 + 防重复
  - 采集后自动评估（collector AlertHook）+ load5 动态阈值
  - [ ] ssh_fail 连续失败计数（当前简化为失败即触发/成功即恢复）
- [x] **阶段六：通知系统（飞书）** — 通道 CRUD + 飞书 webhook + 文案 + 防重复 + 告警接入
  - notification_channels / alert_notifications 模型 + 迁移
  - 飞书 webhook（含可选签名校验）+ 告警/恢复/离线文案（PRD 16）
  - Notifier 接入告警事件：firing 按 notify_interval 防重复，recovered 发恢复通知
  - [ ] Telegram Bot / SMTP 邮件（本次只做飞书）
- [x] **阶段七：仪表盘** — summary/Top5/recent-alerts 真实聚合，端到端验证通过
  - GET /api/dashboard/summary（服务器状态汇总）
  - GET /api/dashboard/top-cpu / top-memory / top-disk（每台最新 metric Top5）
  - GET /api/dashboard/recent-alerts（基于 alert_events）
- [x] **阶段八：前端后台** — Vue3 + TypeScript + Vite + Naive UI + ECharts，10 个页面
  - 全部页面已切换为真实接口（登录/布局/服务器/分组/详情/告警规则/事件/通知通道/仪表盘）

---

## 🚀 部署与运维
- [x] Go embed 前端静态资源（单二进制 ~40M，SPA history 兜底）— 已验证
- [x] 生产配置模板 `configs/config.prod.yaml` + systemd `deploy/gosee.service`
- [x] Nginx 反代 `deploy/nginx.conf` + 完整部署文档 `DEPLOY.md`
- [ ] Dockerfile / docker-compose（单二进制方案已够用，Docker 作为可选增强）
- [ ] MySQL 切换实测（driver=mysql，DEPLOY.md 已写步骤）
- [ ] Swagger / OpenAPI 接口文档

---

## 🛠 技术债 / 优化
- [ ] SSH host key 校验（替代 `InsecureIgnoreHostKey`，PRD 15）
- [ ] 私钥 passphrase 采集时透传支持
- [ ] ssh_fail 连续失败计数（失败历史记录，替代当前的"失败即触发"简化）
- [ ] Telegram Bot / SMTP 邮件通知（阶段六只做了飞书）
- [ ] 指标数据保留策略（定期清理老旧 metrics，避免膨胀）
- [ ] 登录失败次数限制（PRD 15.3）
- [ ] 操作日志（PRD 15.3 / 7.3）
- [ ] `server_metrics` 大数据量下的索引/分区优化
- [ ] 健康检查增强（DB / Redis 连通性）
- [ ] 采集结果重试（`max_retries`，目前仅预留）
- [ ] 服务器列表显示实时 CPU/内存/磁盘列（后端 `/servers` 关联最新 metric 返回，对齐 PRD 列表设计；当前实时指标只在详情页展示）
