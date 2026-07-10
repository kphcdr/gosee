import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

function formatBuildInfo(now: Date) {
  const date = `${now.getFullYear()}.${pad(now.getMonth() + 1)}.${pad(now.getDate())}`
  const time = `${pad(now.getHours())}${pad(now.getMinutes())}`
  const offsetMinutes = -now.getTimezoneOffset()
  const offsetSign = offsetMinutes >= 0 ? '+' : '-'
  const offset = `${offsetSign}${pad(Math.floor(Math.abs(offsetMinutes) / 60))}:${pad(Math.abs(offsetMinutes) % 60)}`
  return {
    version: process.env.VITE_APP_VERSION?.trim() || `v${date}-${time}`,
    buildTime: `${date.replaceAll('.', '-')} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())} UTC${offset}`,
  }
}

const buildInfo = formatBuildInfo(new Date())

// https://vite.dev/config/
export default defineConfig({
  define: {
    __APP_VERSION__: JSON.stringify(buildInfo.version),
    __APP_BUILD_TIME__: JSON.stringify(buildInfo.buildTime),
  },
  plugins: [
    vue(),
    Components({
      resolvers: [NaiveUiResolver()],
      dirs: [], // 只自动导入 naive-ui 组件；自定义组件仍显式 import，更可控
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    // 开发期把 /api 转发到后端，与生产同源部署行为一致，避免 CORS/cookie 差异
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    target: 'es2020',
    chunkSizeWarningLimit: 1200,
  },
})
