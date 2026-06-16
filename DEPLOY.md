# gosee 部署指南（单二进制 + SQLite + systemd）

gosee 采用 **Go embed** 把前端打进后端二进制，部署只需一个可执行文件 + 一个配置文件，零外部依赖（SQLite 文件存储）。

---

## 一、构建产物

项目根目录一键构建（推荐）：
```bash
make all    # = pnpm build + go build，生成单二进制 gosee（~40M，含前端 + 后端）
```

**部署到 Linux**（从 macOS 交叉编译，无需 Linux 编译环境）：
```bash
make build-linux    # → gosee-linux-amd64（静态链接 ELF x86-64，~41M，零依赖）
```
> `glebarez/sqlite` 纯 Go 免 CGO，交叉编译无需 C 工具链。ARM64 服务器把 Makefile 里的 `GOARCH=amd64` 改成 `arm64` 即可。

上传后在服务器上重命名为 `gosee` 即可（后续步骤统一用 `gosee`）。

或手动分步：
```bash
cd web && pnpm install && pnpm build   # 前端 → web/dist
cd .. && go build -o gosee ./cmd/server # 后端（自动 embed 前端）
```

> 顺序很重要：必须先 `pnpm build` 生成 `web/dist`，再 `go build`（否则 embed 报错）。

---

## 二、部署到服务器

### 1. 准备目录与用户
```bash
sudo useradd -r -s /sbin/nologin gosee
sudo mkdir -p /opt/gosee
sudo chown gosee:gosee /opt/gosee
```

### 2. 上传文件
本地执行：
```bash
scp gosee                     deploy@server:/tmp/
scp configs/config.prod.yaml  deploy@server:/tmp/config.yaml
scp deploy/gosee.service      deploy@server:/tmp/
```
服务器执行：
```bash
sudo mv /tmp/gosee        /opt/gosee/gosee
sudo mv /tmp/config.yaml  /opt/gosee/config.yaml
sudo chmod +x /opt/gosee/gosee
sudo chown -R gosee:gosee /opt/gosee
```

### 3. ★必做：修改配置密钥★
```bash
cd /opt/gosee
# 生成两个密钥
openssl rand -hex 32      # → 填入 jwt.secret
openssl rand -hex 32      # → 填入 security.encryption_key
sudo -u gosee vi config.yaml   # 替换所有 CHANGE_ME；改 admin 密码
```
> `security.encryption_key` 不改 = SSH 私钥/密码等于明文存储！

### 4. systemd 守护
```bash
sudo mv /tmp/gosee.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gosee
sudo systemctl status gosee       # 确认 active (running)
```

### 5. 验证
```bash
curl http://localhost:8080/health        # {"status":"ok"}
curl http://localhost:8080/              # 返回前端首页 HTML
```
浏览器访问 `http://服务器IP:8080`，用配置的 admin 账号登录。

### 6. （可选）Nginx 反代 + HTTPS
```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/gosee
# 改 server_name 为你的域名
sudo ln -s /etc/nginx/sites-available/gosee /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
# 上 HTTPS：
sudo certbot --nginx -d monitor.example.com
```
> gosee 已 embed 前端，Nginx 只需反代到 8080（前端 + API 同源），无需单独托管静态文件。

---

## 三、网络要求

- **部署机 → 被监控机**：SSH 可达（22 或自定义端口）—— 这是 SSH 拉取式监控的前提
- **用户 → 部署机**：开放 80/443（Nginx）或 8080（直连）
- 防火墙建议只开 22 / 80 / 443

---

## 四、运维

### 查看日志
```bash
sudo journalctl -u gosee -f          # systemd 日志
# 或 /opt/gosee/logs/                # 应用日志文件
```

### 升级
```bash
# 本地重新 build（pnpm build && go build）→ scp 新二进制 → 重启
sudo systemctl restart gosee
```

### 备份
```bash
# SQLite（停服或用 .backup 避免锁）
sqlite3 /opt/gosee/gosee.db ".backup /backup/gosee-$(date +%F).db"
# MySQL：mysqldump -u root gosee > backup.sql
```

### 切换到 MySQL（规模增大时）
改 `config.yaml`：
```yaml
database:
  driver: mysql
  dsn: "root:pass@tcp(127.0.0.1:3306)/gosee?charset=utf8mb4&parseTime=True&loc=Local"
```
重启即可，代码无需改动（首次会自动迁移表 + 种子数据）。

---

## 五、生产加固清单

- [x] 单二进制部署（embed 前端）
- [ ] `jwt.secret` 已替换为强随机
- [ ] `security.encryption_key` 已替换为强随机
- [ ] admin 密码已改为强密码
- [ ] Nginx + HTTPS 已启用
- [ ] 防火墙仅开必要端口
- [ ] 定期备份 gosee.db
