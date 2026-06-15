<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import type { DataTableColumns } from 'naive-ui'
import { serverApi } from '@/api/server'
import type { Server } from '@/types/server'
import type { ServerMetric, ServerDisk } from '@/types/metric'
import { formatTime, formatBytes, formatPercent, formatUptime } from '@/utils/format'
import { statusLabel, statusType } from '@/constants/server'
import { message } from '@/utils/discrete'
import { usePolling } from '@/composables/usePolling'
import TrendChart from '@/components/TrendChart.vue'

const route = useRoute()
const id = computed(() => Number(route.params.id))

const server = ref<Server | null>(null)
const metrics = ref<ServerMetric[]>([])
const disks = ref<ServerDisk[]>([])
const loading = ref(false)
const collecting = ref(false)

const hours = ref(24)
const hourOptions = [
  { label: '最近 1 小时', value: 1 },
  { label: '最近 6 小时', value: 6 },
  { label: '最近 24 小时', value: 24 },
  { label: '最近 7 天', value: 168 },
]

async function loadServer() {
  server.value = await serverApi.get(id.value)
}
async function loadMetrics() {
  const res = await serverApi.metrics(id.value, { hours: hours.value, limit: 500 })
  metrics.value = res.list || []
}
async function loadDisks() {
  disks.value = await serverApi.disks(id.value)
}
async function loadAll() {
  loading.value = true
  try {
    await Promise.all([loadServer(), loadMetrics(), loadDisks()])
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
usePolling(loadMetrics, 60_000)

async function onHoursChange() {
  await loadMetrics()
}

async function doCollect() {
  collecting.value = true
  try {
    await serverApi.collect(id.value)
    message.success('采集成功')
    await loadAll()
  } finally {
    collecting.value = false
  }
}

const times = computed(() => metrics.value.map((m) => formatTime(m.collected_at, 'MM-DD HH:mm')))
const cpuSeries = computed(() => [{ name: 'CPU', data: metrics.value.map((m) => m.cpu_usage) }])
const memSeries = computed(() => [{ name: '内存使用率', data: metrics.value.map((m) => m.memory_usage) }])
const loadSeries = computed(() => [
  { name: 'load1', data: metrics.value.map((m) => m.load_1) },
  { name: 'load5', data: metrics.value.map((m) => m.load_5) },
  { name: 'load15', data: metrics.value.map((m) => m.load_15) },
])
const diskSeries = computed(() => [{ name: '磁盘最高', data: metrics.value.map((m) => m.disk_max_usage) }])

const latest = computed(() => (metrics.value.length ? metrics.value[metrics.value.length - 1] : null))

const diskColumns: DataTableColumns<ServerDisk> = [
  { title: '挂载点', key: 'mount_point' },
  { title: '文件系统', key: 'filesystem', ellipsis: { tooltip: true } },
  { title: '总容量', key: 'size_bytes', render: (r) => formatBytes(r.size_bytes) },
  { title: '已用', key: 'used_bytes', render: (r) => formatBytes(r.used_bytes) },
  { title: '可用', key: 'available_bytes', render: (r) => formatBytes(r.available_bytes) },
  { title: '使用率', key: 'usage_percent', render: (r) => formatPercent(r.usage_percent) },
]
</script>

<template>
  <n-spin :show="loading">
    <n-space vertical size="large">
      <n-card v-if="server" title="基础信息">
        <template #header-extra>
          <n-space>
            <n-select
              v-model:value="hours"
              :options="hourOptions"
              style="width: 150px"
              @update:value="onHoursChange"
            />
            <n-button type="primary" :loading="collecting" @click="doCollect">手动采集</n-button>
          </n-space>
        </template>
        <n-descriptions :column="4" label-placement="left" bordered>
          <n-descriptions-item label="名称">{{ server.name }}</n-descriptions-item>
          <n-descriptions-item label="主机">{{ server.host }}:{{ server.port }}</n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag :type="statusType(server.status)" size="small" round :bordered="false">
              {{ statusLabel(server.status) }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="登录用户">{{ server.username }}</n-descriptions-item>
          <n-descriptions-item label="系统">{{ latest?.os || '-' }}</n-descriptions-item>
          <n-descriptions-item label="主机名">{{ latest?.hostname || '-' }}</n-descriptions-item>
          <n-descriptions-item label="运行时长">{{ formatUptime(latest?.uptime_seconds) }}</n-descriptions-item>
          <n-descriptions-item label="最后采集">{{ formatTime(server.last_checked_at) }}</n-descriptions-item>
        </n-descriptions>
      </n-card>

      <n-grid v-if="latest" :cols="4" :x-gap="16">
        <n-gi>
          <n-card><n-statistic label="CPU 使用率" :value="formatPercent(latest.cpu_usage)" /></n-card>
        </n-gi>
        <n-gi>
          <n-card><n-statistic label="内存使用率" :value="formatPercent(latest.memory_usage)" /></n-card>
        </n-gi>
        <n-gi>
          <n-card><n-statistic label="磁盘最高" :value="formatPercent(latest.disk_max_usage)" /></n-card>
        </n-gi>
        <n-gi>
          <n-card>
            <n-statistic label="负载 (1/5/15)" :value="`${latest.load_1} / ${latest.load_5} / ${latest.load_15}`" />
          </n-card>
        </n-gi>
      </n-grid>

      <n-grid :cols="2" :x-gap="16" :y-gap="16">
        <n-gi>
          <n-card title="CPU 使用率">
            <TrendChart :x-data="times" :series="cpuSeries" unit="%" :y-max="100" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card title="内存使用率">
            <TrendChart :x-data="times" :series="memSeries" unit="%" :y-max="100" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card title="系统负载">
            <TrendChart :x-data="times" :series="loadSeries" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card title="磁盘最高使用率">
            <TrendChart :x-data="times" :series="diskSeries" unit="%" :y-max="100" />
          </n-card>
        </n-gi>
      </n-grid>

      <n-card title="磁盘明细">
        <n-data-table :columns="diskColumns" :data="disks" :bordered="false" size="small" />
      </n-card>
    </n-space>
  </n-spin>
</template>
