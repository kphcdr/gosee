<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import { NTag, type DataTableColumns } from 'naive-ui'
import { dashboardApi } from '@/api/dashboard'
import { serverGroupApi } from '@/api/serverGroup'
import type { DashboardSummary, TopItem, RecentAlert } from '@/types/dashboard'
import { formatTime, formatPercent } from '@/utils/format'
import { levelLabel, levelType } from '@/constants/alert'
import { usePolling } from '@/composables/usePolling'
import RankBars from '@/components/RankBars.vue'
import type { ServerGroup } from '@/types/serverGroup'

const summary = ref<DashboardSummary | null>(null)
const topCpu = ref<TopItem[]>([])
const topMemory = ref<TopItem[]>([])
const topDisk = ref<TopItem[]>([])
const recentAlerts = ref<RecentAlert[]>([])
const loading = ref(false)
const groups = ref<ServerGroup[]>([])
const selectedGroupId = ref(0)
const groupOptions = computed(() => [
  { label: '全部分组', value: 0 },
  ...groups.value.map((group) => ({ label: group.name, value: group.id })),
])

async function loadAll() {
  loading.value = true
  try {
    const groupId = selectedGroupId.value || undefined
    const [s, c, m, d, a] = await Promise.all([
      dashboardApi.summary(groupId),
      dashboardApi.topCpu(groupId),
      dashboardApi.topMemory(groupId),
      dashboardApi.topDisk(groupId),
      dashboardApi.recentAlerts(groupId),
    ])
    summary.value = s
    topCpu.value = c
    topMemory.value = m
    topDisk.value = d
    recentAlerts.value = a
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  groups.value = await serverGroupApi.list().catch(() => [])
  await loadAll()
})
usePolling(loadAll, 30_000)

const alertColumns: DataTableColumns<RecentAlert> = [
  { title: '服务器', key: 'server_name' },
  { title: '指标', key: 'metric' },
  { title: '当前值', key: 'value', render: (r) => formatPercent(r.value) },
  {
    title: '级别',
    key: 'level',
    width: 90,
    render: (r) =>
      h(
        NTag,
        { type: levelType(r.level), size: 'small', round: true, bordered: false },
        { default: () => levelLabel(r.level) },
      ),
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (r) =>
      h(
        NTag,
        {
          type: r.status === 'firing' ? 'error' : 'success',
          size: 'small',
          round: true,
          bordered: false,
        },
        { default: () => (r.status === 'firing' ? '告警中' : '已恢复') },
      ),
  },
  { title: '触发时间', key: 'triggered_at', render: (r) => formatTime(r.triggered_at) },
]
</script>

<template>
  <n-spin :show="loading">
    <n-space vertical size="large">
      <n-card size="small">
        <n-space align="center">
          <span class="filter-label">服务器分组</span>
          <n-select
            v-model:value="selectedGroupId"
            :options="groupOptions"
            style="width: 220px"
            @update:value="loadAll"
          />
        </n-space>
      </n-card>

      <n-grid :cols="4" :x-gap="16">
        <n-gi>
          <n-card><n-statistic label="服务器总数" :value="summary?.total ?? 0" /></n-card>
        </n-gi>
        <n-gi>
          <n-card><n-statistic label="正常" :value="summary?.normal ?? 0" /></n-card>
        </n-gi>
        <n-gi>
          <n-card>
            <n-statistic label="告警" :value="(summary?.warning ?? 0) + (summary?.critical ?? 0)" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card><n-statistic label="离线" :value="summary?.offline ?? 0" /></n-card>
        </n-gi>
      </n-grid>

      <n-grid :cols="3" :x-gap="16">
        <n-gi><RankBars :items="topCpu" title="CPU 使用率 Top 5" /></n-gi>
        <n-gi><RankBars :items="topMemory" title="内存使用率 Top 5" /></n-gi>
        <n-gi><RankBars :items="topDisk" title="磁盘使用率 Top 5" /></n-gi>
      </n-grid>

      <n-card title="最近告警">
        <n-data-table :columns="alertColumns" :data="recentAlerts" :bordered="false" size="small" />
      </n-card>
    </n-space>
  </n-spin>
</template>

<style scoped>
.filter-label {
  color: var(--n-text-color);
  font-weight: 500;
  white-space: nowrap;
}
</style>
