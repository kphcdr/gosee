<script setup lang="ts">
import { ref, h, onMounted, type VNode } from 'vue'
import { NButton, NSpace, NTag, type DataTableColumns } from 'naive-ui'
import { alertEventApi } from '@/api/alertEvent'
import type { AlertEvent } from '@/types/alert'
import { levelLabel, levelType } from '@/constants/alert'
import { formatTime } from '@/utils/format'
import { message } from '@/utils/discrete'
import { usePolling } from '@/composables/usePolling'

const loading = ref(false)
const list = ref<AlertEvent[]>([])
const statusFilter = ref<string | null>(null)

const statusOptions = [
  { label: '全部', value: '' },
  { label: '告警中', value: 'firing' },
  { label: '已确认', value: 'acked' },
  { label: '已关闭', value: 'closed' },
  { label: '已恢复', value: 'recovered' },
]

async function load() {
  loading.value = true
  try {
    const all = await alertEventApi.list()
    if (statusFilter.value === 'acked') {
      list.value = all.filter((e) => e.acked_at !== null)
    } else {
      list.value = statusFilter.value ? all.filter((e) => e.status === statusFilter.value) : all
    }
  } finally {
    loading.value = false
  }
}
onMounted(load)
usePolling(load, 30_000)

type TagType = 'default' | 'info' | 'success' | 'warning' | 'error'

function statusMeta(event: AlertEvent): { label: string; type: TagType } {
  if (event.status === 'firing' && event.acked_at) {
    return { label: '告警中（已确认）', type: 'warning' }
  }
  const map: Record<string, { label: string; type: TagType }> = {
    firing: { label: '告警中', type: 'error' },
    closed: { label: '已关闭', type: 'default' },
    recovered: { label: '已恢复', type: 'success' },
  }
  return map[event.status] || { label: event.status, type: 'default' }
}

async function doAck(row: AlertEvent) {
  await alertEventApi.ack(row.id)
  message.success('已确认')
  await load()
}
async function doClose(row: AlertEvent) {
  await alertEventApi.close(row.id)
  message.success('已关闭')
  await load()
}

const columns: DataTableColumns<AlertEvent> = [
  { title: '服务器', key: 'server_name', width: 120 },
  { title: '规则', key: 'rule_name', ellipsis: { tooltip: true } },
  { title: '指标', key: 'metric', width: 120 },
  { title: '当前值', key: 'current_value', width: 90 },
  { title: '阈值', key: 'threshold', width: 80 },
  {
    title: '级别',
    key: 'level',
    width: 80,
    render: (r) =>
      h(NTag, { type: levelType(r.level), size: 'small', bordered: false }, { default: () => levelLabel(r.level) }),
  },
  {
    title: '状态',
    key: 'status',
    width: 150,
    render: (r) => {
      const m = statusMeta(r)
      return h(NTag, { type: m.type, size: 'small', round: true, bordered: false }, { default: () => m.label })
    },
  },
  { title: '首次触发', key: 'first_triggered_at', width: 170, render: (r) => formatTime(r.first_triggered_at) },
  { title: '最近触发', key: 'last_triggered_at', width: 170, render: (r) => formatTime(r.last_triggered_at) },
  { title: '恢复时间', key: 'recovered_at', width: 170, render: (r) => formatTime(r.recovered_at) },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    render: (r) =>
      h(NSpace, { size: 'small' }, {
        default: (): VNode | VNode[] => {
          const btns: VNode[] = []
          if (r.status === 'firing' && !r.acked_at) {
            btns.push(h(NButton, { size: 'small', onClick: () => doAck(r) }, { default: () => '确认' }))
          }
          if (r.status === 'firing') {
            btns.push(
              h(NButton, { size: 'small', type: 'warning', onClick: () => doClose(r) }, { default: () => '关闭' }),
            )
          }
          return btns.length ? btns : h('span', { style: 'color:#bbb' }, '-')
        },
      }),
  },
]
</script>

<template>
  <n-card>
    <div class="flex items-center justify-between mb-4">
      <n-select
        v-model:value="statusFilter"
        :options="statusOptions"
        placeholder="状态筛选"
        style="width: 160px"
        @update:value="load"
      />
    </div>
    <n-data-table :columns="columns" :data="list" :loading="loading" :bordered="false" size="small" />
  </n-card>
</template>
