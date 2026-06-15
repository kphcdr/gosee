<script setup lang="ts">
import { ref, reactive, onMounted, computed, h } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NSpace, NTag, type DataTableColumns, type PaginationProps } from 'naive-ui'
import { serverApi } from '@/api/server'
import { serverGroupApi } from '@/api/serverGroup'
import type { Server, ServerListQuery } from '@/types/server'
import type { ServerGroup } from '@/types/serverGroup'
import type { PagedResult } from '@/types/api'
import { formatTime } from '@/utils/format'
import { statusLabel, statusType, ENABLED_OPTIONS } from '@/constants/server'
import { message, dialog } from '@/utils/discrete'
import { usePolling } from '@/composables/usePolling'
import ServerForm from './components/ServerForm.vue'

const router = useRouter()
const loading = ref(false)
const list = ref<Server[]>([])
const groups = ref<ServerGroup[]>([])

const query = reactive<ServerListQuery>({
  page: 1,
  page_size: 20,
  keyword: '',
  group_id: undefined,
  enabled: undefined,
})

const groupOptions = computed(() => groups.value.map((g) => ({ label: g.name, value: g.id })))

const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  prefix: ({ itemCount }) => `共 ${itemCount} 条`,
})

async function load() {
  loading.value = true
  try {
    const res: PagedResult<Server> = await serverApi.list({ ...query })
    list.value = res.list || []
    pagination.itemCount = res.total
    pagination.page = query.page || 1
    pagination.pageSize = query.page_size || 20
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  groups.value = await serverGroupApi.list().catch(() => [])
}

onMounted(async () => {
  await loadGroups()
  await load()
})

usePolling(load, 60_000)

function onSearch() {
  query.page = 1
  load()
}
function onReset() {
  query.keyword = ''
  query.group_id = undefined
  query.enabled = undefined
  query.page = 1
  load()
}
function onPageChange(page: number) {
  query.page = page
  load()
}
function onPageSizeChange(ps: number) {
  query.page_size = ps
  query.page = 1
  load()
}

function groupName(id: number | null | undefined): string {
  if (id == null) return '-'
  return groups.value.find((g) => g.id === id)?.name || `#${id}`
}

const testingId = ref<number | null>(null)
const collectingId = ref<number | null>(null)

async function doTestSSH(id: number) {
  testingId.value = id
  try {
    await serverApi.testSSH(id)
    message.success('SSH 连接正常')
    await load()
  } finally {
    testingId.value = null
  }
}

async function doCollect(id: number) {
  collectingId.value = id
  try {
    await serverApi.collect(id)
    message.success('采集成功')
    await load()
  } finally {
    collectingId.value = null
  }
}

function confirmDelete(row: Server) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除服务器「${row.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await serverApi.remove(row.id)
      message.success('删除成功')
      await load()
    },
  })
}

const columns: DataTableColumns<Server> = [
  { title: '名称', key: 'name', ellipsis: { tooltip: true } },
  { title: 'IP', key: 'host', width: 170, render: (r) => `${r.host}:${r.port}` },
  { title: '分组', key: 'group_id', width: 120, render: (r) => groupName(r.group_id) },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (r) =>
      h(
        NTag,
        { type: statusType(r.status), size: 'small', round: true, bordered: false },
        { default: () => statusLabel(r.status) },
      ),
  },
  { title: '最后采集', key: 'last_checked_at', width: 170, render: (r) => formatTime(r.last_checked_at) },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    render: (r) =>
      h(
        NSpace,
        { size: 'small', wrap: false },
        {
          default: () => [
            h(NButton, { size: 'small', onClick: () => router.push(`/servers/${r.id}`) }, { default: () => '详情' }),
            h(NButton, { size: 'small', onClick: () => openEdit(r) }, { default: () => '编辑' }),
            h(
              NButton,
              { size: 'small', loading: testingId.value === r.id, onClick: () => doTestSSH(r.id) },
              { default: () => '测试' },
            ),
            h(
              NButton,
              { size: 'small', loading: collectingId.value === r.id, onClick: () => doCollect(r.id) },
              { default: () => '采集' },
            ),
            h(
              NButton,
              { size: 'small', type: 'error', ghost: true, onClick: () => confirmDelete(r) },
              { default: () => '删除' },
            ),
          ],
        },
      ),
  },
]

const showForm = ref(false)
const editing = ref<Server | null>(null)
function openCreate() {
  editing.value = null
  showForm.value = true
}
function openEdit(row: Server) {
  editing.value = row
  showForm.value = true
}
function onSaved() {
  showForm.value = false
  load()
}
</script>

<template>
  <n-card>
    <div class="flex items-center justify-between mb-4">
      <n-space>
        <n-input
          v-model:value="query.keyword"
          placeholder="名称/IP"
          clearable
          style="width: 170px"
          @keyup.enter="onSearch"
        />
        <n-select
          v-model:value="query.group_id"
          :options="groupOptions"
          clearable
          placeholder="分组"
          style="width: 140px"
          @update:value="onSearch"
        />
        <n-select
          v-model:value="query.enabled"
          :options="ENABLED_OPTIONS"
          clearable
          placeholder="启用"
          style="width: 110px"
          @update:value="onSearch"
        />
        <n-button @click="onSearch">搜索</n-button>
        <n-button @click="onReset">重置</n-button>
      </n-space>
      <n-button type="primary" @click="openCreate">+ 新增服务器</n-button>
    </div>
    <n-data-table
      remote
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      :bordered="false"
      :row-key="(row: Server) => row.id"
      @update:page="onPageChange"
      @update:page-size="onPageSizeChange"
    />
  </n-card>

  <ServerForm v-model:show="showForm" :model="editing" @saved="onSaved" />
</template>
