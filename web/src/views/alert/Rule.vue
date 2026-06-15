<script setup lang="ts">
import { ref, reactive, onMounted, computed, h } from 'vue'
import { NButton, NSpace, NSwitch, NTag, type DataTableColumns, type FormInst, type FormRules } from 'naive-ui'
import { alertRuleApi } from '@/api/alertRule'
import { serverApi } from '@/api/server'
import { serverGroupApi } from '@/api/serverGroup'
import type { AlertRule } from '@/types/alert'
import type { Server } from '@/types/server'
import type { ServerGroup } from '@/types/serverGroup'
import {
  METRIC_TYPE_OPTIONS,
  OPERATOR_OPTIONS,
  LEVEL_OPTIONS,
  SCOPE_TYPE_OPTIONS,
  levelLabel,
  levelType,
  type MetricType,
  type ScopeType,
  type AlertLevel,
} from '@/constants/alert'
import { message, dialog } from '@/utils/discrete'

const loading = ref(false)
const list = ref<AlertRule[]>([])

async function load() {
  loading.value = true
  try {
    list.value = await alertRuleApi.list()
  } finally {
    loading.value = false
  }
}
onMounted(load)

function metricLabel(v: string): string {
  return METRIC_TYPE_OPTIONS.find((o) => o.value === v)?.label || v
}
function scopeLabel(v: string): string {
  return SCOPE_TYPE_OPTIONS.find((o) => o.value === v)?.label || v
}

async function toggleEnabled(row: AlertRule, val: boolean) {
  try {
    if (val) await alertRuleApi.enable(row.id)
    else await alertRuleApi.disable(row.id)
    message.success(val ? '已启用' : '已禁用')
    await load()
  } catch {
    // 错误已由拦截器提示
  }
}

const columns: DataTableColumns<AlertRule> = [
  { title: '名称', key: 'name', ellipsis: { tooltip: true } },
  { title: '指标', key: 'metric_type', render: (r) => metricLabel(r.metric_type) },
  { title: '条件', key: 'cond', width: 110, render: (r) => `${r.operator} ${r.threshold}` },
  { title: '连续(次)', key: 'duration_count', width: 90 },
  {
    title: '级别',
    key: 'level',
    width: 80,
    render: (r) =>
      h(NTag, { type: levelType(r.level), size: 'small', bordered: false }, { default: () => levelLabel(r.level) }),
  },
  { title: '范围', key: 'scope_type', width: 130, render: (r) => scopeLabel(r.scope_type) },
  {
    title: '启用',
    key: 'enabled',
    width: 70,
    render: (r) => h(NSwitch, { value: r.enabled === 1, size: 'small', onUpdateValue: (v: boolean) => toggleEnabled(r, v) }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    render: (r) =>
      h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => openEdit(r) }, { default: () => '编辑' }),
          h(
            NButton,
            { size: 'small', type: 'error', ghost: true, onClick: () => confirmDelete(r) },
            { default: () => '删除' },
          ),
        ],
      }),
  },
]

// 表单
const showForm = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInst>()
const saving = ref(false)
const groups = ref<ServerGroup[]>([])
const servers = ref<Server[]>([])

const form = reactive({
  name: '',
  metric_type: 'cpu_usage' as MetricType,
  operator: '>',
  threshold: 90,
  duration_count: 3,
  level: 'warning' as AlertLevel,
  scope_type: 'global' as ScopeType,
  scope_id: null as number | null,
  enabled: 1 as 0 | 1,
})

const rules: FormRules = {
  name: { required: true, message: '请输入规则名称', trigger: 'blur' },
}

const scopeTargetOptions = computed(() => {
  if (form.scope_type === 'group') return groups.value.map((g) => ({ label: g.name, value: g.id }))
  if (form.scope_type === 'server') return servers.value.map((s) => ({ label: s.name, value: s.id }))
  return []
})

async function loadTargets() {
  const [g, s] = await Promise.all([
    serverGroupApi.list().catch(() => []),
    serverApi
      .list({ page: 1, page_size: 999 })
      .then((r) => r.list)
      .catch(() => []),
  ])
  groups.value = g
  servers.value = s
}

async function openCreate() {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, {
    name: '',
    metric_type: 'cpu_usage',
    operator: '>',
    threshold: 90,
    duration_count: 3,
    level: 'warning',
    scope_type: 'global',
    scope_id: null,
    enabled: 1,
  })
  await loadTargets()
  showForm.value = true
}

async function openEdit(row: AlertRule) {
  isEdit.value = true
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    metric_type: row.metric_type,
    operator: row.operator,
    threshold: row.threshold,
    duration_count: row.duration_count,
    level: row.level,
    scope_type: row.scope_type,
    scope_id: row.scope_id,
    enabled: row.enabled,
  })
  await loadTargets()
  showForm.value = true
}

async function save() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const body = { ...form }
    if (form.scope_type === 'global') body.scope_id = null
    if (isEdit.value && editingId.value != null) {
      await alertRuleApi.update(editingId.value, body)
    } else {
      await alertRuleApi.create(body)
    }
    message.success('保存成功')
    showForm.value = false
    await load()
  } finally {
    saving.value = false
  }
}

function confirmDelete(row: AlertRule) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除规则「${row.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await alertRuleApi.remove(row.id)
      message.success('删除成功')
      await load()
    },
  })
}
</script>

<template>
  <n-card>
    <div class="flex items-center justify-between mb-4">
      <span style="font-size: 15px; font-weight: 500">告警规则</span>
      <n-button type="primary" @click="openCreate">+ 新增规则</n-button>
    </div>
    <n-data-table :columns="columns" :data="list" :loading="loading" :bordered="false" />
  </n-card>

  <n-modal v-model:show="showForm" preset="card" :title="isEdit ? '编辑规则' : '新增规则'" style="width: 540px">
    <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="90">
      <n-form-item label="规则名称" path="name">
        <n-input v-model:value="form.name" placeholder="如 CPU 使用率过高" />
      </n-form-item>
      <n-form-item label="监控指标">
        <n-select v-model:value="form.metric_type" :options="METRIC_TYPE_OPTIONS" />
      </n-form-item>
      <n-grid :cols="2" :x-gap="12">
        <n-form-item-gi label="运算符">
          <n-select v-model:value="form.operator" :options="OPERATOR_OPTIONS" />
        </n-form-item-gi>
        <n-form-item-gi label="阈值" path="threshold">
          <n-input-number
            :value="form.threshold"
            :min="0"
            style="width: 100%"
            @update:value="(v: number | null) => (form.threshold = v ?? 0)"
          />
        </n-form-item-gi>
      </n-grid>
      <n-grid :cols="2" :x-gap="12">
        <n-form-item-gi label="连续触发">
          <n-input-number
            :value="form.duration_count"
            :min="1"
            style="width: 100%"
            @update:value="(v: number | null) => (form.duration_count = v ?? 1)"
          />
        </n-form-item-gi>
        <n-form-item-gi label="告警级别">
          <n-select v-model:value="form.level" :options="LEVEL_OPTIONS" />
        </n-form-item-gi>
      </n-grid>
      <n-form-item label="生效范围">
        <n-select v-model:value="form.scope_type" :options="SCOPE_TYPE_OPTIONS" />
      </n-form-item>
      <n-form-item v-if="form.scope_type !== 'global'" :label="form.scope_type === 'group' ? '选择分组' : '选择服务器'">
        <n-select v-model:value="form.scope_id" :options="scopeTargetOptions" clearable placeholder="请选择" />
      </n-form-item>
      <n-form-item label="启用">
        <n-switch :value="form.enabled === 1" @update:value="(v: boolean) => (form.enabled = v ? 1 : 0)" />
      </n-form-item>
    </n-form>
    <template #footer>
      <div class="flex justify-between">
        <n-button @click="showForm = false">取消</n-button>
        <n-button type="primary" :loading="saving" @click="save">保存</n-button>
      </div>
    </template>
  </n-modal>
</template>
