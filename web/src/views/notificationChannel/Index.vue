<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { NButton, NSpace, NTag, type DataTableColumns, type FormInst, type FormRules } from 'naive-ui'
import { notificationChannelApi } from '@/api/notificationChannel'
import type { NotificationChannel, ChannelType } from '@/types/notification'
import { CHANNEL_TYPE_OPTIONS, CHANNEL_FIELDS } from '@/constants/notification'
import { formatTime } from '@/utils/format'
import { message, dialog } from '@/utils/discrete'

const loading = ref(false)
const list = ref<NotificationChannel[]>([])
const testingId = ref<number | null>(null)

async function load() {
  loading.value = true
  try {
    list.value = await notificationChannelApi.list()
  } finally {
    loading.value = false
  }
}
onMounted(load)

function typeLabel(v: string): string {
  return CHANNEL_TYPE_OPTIONS.find((o) => o.value === v)?.label || v
}

const columns: DataTableColumns<NotificationChannel> = [
  { title: '名称', key: 'name' },
  {
    title: '类型',
    key: 'type',
    width: 140,
    render: (r) => h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => typeLabel(r.type) }),
  },
  { title: '启用', key: 'enabled', width: 70, render: (r) => (r.enabled === 1 ? '是' : '否') },
  { title: '创建时间', key: 'created_at', width: 170, render: (r) => formatTime(r.created_at) },
  {
    title: '操作',
    key: 'actions',
    width: 230,
    render: (r) =>
      h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => openEdit(r) }, { default: () => '编辑' }),
          h(
            NButton,
            { size: 'small', type: 'info', ghost: true, loading: testingId.value === r.id, onClick: () => doTest(r.id) },
            { default: () => '测试' },
          ),
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

const form = reactive<{
  name: string
  type: ChannelType
  config: Record<string, string>
  enabled: 0 | 1
}>({
  name: '',
  type: 'feishu',
  config: {},
  enabled: 1,
})

const rules: FormRules = {
  name: { required: true, message: '请输入通道名称', trigger: 'blur' },
}

const fields = computed(() => CHANNEL_FIELDS[form.type] || [])

function initConfig() {
  const oldConfig = { ...form.config }
  form.config = {}
  CHANNEL_FIELDS[form.type].forEach((f) => {
    form.config[f.key] = oldConfig[f.key] ?? ''
  })
}

function openCreate() {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, { name: '', type: 'feishu', config: {}, enabled: 1 })
  initConfig()
  showForm.value = true
}
function openEdit(row: NotificationChannel) {
  isEdit.value = true
  editingId.value = row.id
  Object.assign(form, { name: row.name, type: row.type, config: { ...row.config }, enabled: row.enabled })
  initConfig()
  showForm.value = true
}

function onTypeChange() {
  initConfig()
}

async function save() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const body = {
      name: form.name,
      type: form.type,
      config: { ...form.config },
      enabled: form.enabled,
    }
    if (isEdit.value && editingId.value != null) {
      await notificationChannelApi.update(editingId.value, body)
    } else {
      await notificationChannelApi.create(body)
    }
    message.success('保存成功')
    showForm.value = false
    await load()
  } finally {
    saving.value = false
  }
}

function confirmDelete(row: NotificationChannel) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除通道「${row.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await notificationChannelApi.remove(row.id)
      message.success('删除成功')
      await load()
    },
  })
}

async function doTest(id: number) {
  testingId.value = id
  try {
    await notificationChannelApi.test(id)
    message.success('测试消息已发送')
  } finally {
    testingId.value = null
  }
}
</script>

<template>
  <n-card>
    <div class="flex items-center justify-between mb-4">
      <span style="font-size: 15px; font-weight: 500">通知通道</span>
      <n-button type="primary" @click="openCreate">+ 新增通道</n-button>
    </div>
    <n-data-table :columns="columns" :data="list" :loading="loading" :bordered="false" />
  </n-card>

  <n-modal v-model:show="showForm" preset="card" :title="isEdit ? '编辑通道' : '新增通道'" style="width: 520px">
    <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="110">
      <n-form-item label="通道名称" path="name">
        <n-input v-model:value="form.name" placeholder="如 运维飞书群" />
      </n-form-item>
      <n-form-item label="通道类型">
        <n-select v-model:value="form.type" :options="CHANNEL_TYPE_OPTIONS" @update:value="onTypeChange" />
      </n-form-item>
      <n-form-item v-for="f in fields" :key="f.key" :label="f.label">
        <n-input
          :value="form.config[f.key]"
          :type="f.secret ? 'password' : 'text'"
          :show-password-on="f.secret ? 'click' : undefined"
          :placeholder="f.placeholder"
          @update:value="(v: string) => (form.config[f.key] = v)"
        />
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
