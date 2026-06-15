package web

import "embed"

// Dist 前端构建产物（cd web && pnpm build 生成 web/dist）。
// go build 时把整个 dist 嵌入二进制，实现单文件部署（前端 + 后端一体）。
//
//go:embed all:dist
var Dist embed.FS
