# gosee 前端后台

Vue3 + TypeScript + Vite + Naive UI + ECharts + Pinia + Vue Router + unplugin-vue-components（Naive 组件按需自动导入）。

## 开发

```bash
pnpm install
pnpm dev      # http://localhost:5173，自动代理 /api → 后端 :8080
pnpm build    # 输出 dist/（由后端 Go embed 打进二进制）
```

开发时需同时启动后端（项目根 `make run` 或 `go run ./cmd/server`），登录 `admin / admin123`。

## 技术栈

| 库 | 用途 |
| --- | --- |
| vue / vue-router / pinia | 框架 / 路由 / 状态 |
| naive-ui | UI 组件库 |
| echarts | 趋势图（自封装 `TrendChart`，按需引入） |
| axios | HTTP（拦截器统一加 token、剥壳、401 跳登录） |
| @vueuse/core | 轮询（useIntervalFn） |
| @vicons/ionicons5 | 图标 |
| dayjs | 时间格式化 |

## 目录结构

```
src/
├── api/            # 接口封装（http.ts 拦截器 + 各资源 API）
├── views/          # 页面
│   ├── login/ dashboard/
│   ├── server/     # List / Detail / components/ServerForm
│   ├── serverGroup/
│   ├── alert/      # Rule / Event
│   └── notificationChannel/
├── components/     # 复用组件（TrendChart / RankBars / ChangePasswordModal）
├── composables/    # usePolling（轮询）
├── store/          # Pinia（auth）
├── router/         # 路由 + 守卫
├── types/          # TS 类型（对齐后端模型）
├── constants/      # 枚举（状态色 / 告警级别 / 通知类型）
└── utils/          # discrete(Naive全局) / storage / format / render
```

## 约定

- 统一响应 `{ code, message, data }`，axios 响应拦截器自动剥 `data`，业务层直接拿最终类型
- 401 → 清 token 跳登录（`window.location` 避免 http↔router 循环依赖）
- 路由守卫：无 token 跳 `/login?redirect=`，已登录访问登录页跳仪表盘
- 前端通过 Vite proxy（dev）或 Go embed（prod）与后端同源，无跨域问题
