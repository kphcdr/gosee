<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { NButton, NSpace, type DataTableColumns, type FormInst, type FormRules } from 'naive-ui'
import { serverGroupApi } from '@/api/serverGroup'
import type { ServerGroup, ServerGroupInput } from '@/types/serverGroup'
import { formatTime } from '@/utils/format'
import { message, dialog } from '@/utils/discrete'

const loading = ref(false)
const list = ref<ServerGroup[]>([])
const keyword = ref('')

async function load() {
  loading.value = true
  try {
    list.value = await serverGroupApi.list(keyword.value || undefined)
  } finally {
    loading.value = false
  }
}

onMounted(load)

const columns: DataTableColumns<ServerGroup> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '分组名称', key: 'name' },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true }, render: (r) => r.remark || '-' },
  { title: '创建时间', key: 'created_at', width: 180, render: (r) => formatTime(r.created_at) },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    render: (row) =>
      h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
          h(
            NButton,
            { size: 'small', type: 'error', onClick: () => confirmDelete(row) },
            { default: () => '删除' },
          ),
        ],
      }),
  },
]

// 表单弹窗
const showForm = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInst>()
const saving = ref(false)
const form = reactive<ServerGroupInput>({ name: '', remark: '' })

const rules: FormRules = {
  name: { required: true, message: '请输入分组名称', trigger: 'blur' },
}

function openCreate() {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, { name: '', remark: '' })
  showForm.value = true
}

function openEdit(row: ServerGroup) {
  isEdit.value = true
  editingId.value = row.id
  Object.assign(form, { name: row.name, remark: row.remark })
  showForm.value = true
}

async function save() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const body: ServerGroupInput = { name: form.name, remark: form.remark }
    if (isEdit.value && editingId.value != null) {
      await serverGroupApi.update(editingId.value, body)
    } else {
      await serverGroupApi.create(body)
    }
    message.success('保存成功')
    showForm.value = false
    await load()
  } finally {
    saving.value = false
  }
}

function confirmDelete(row: ServerGroup) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除分组「${row.name}」吗？若分组下仍有服务器将被拒绝。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await serverGroupApi.remove(row.id)
      message.success('删除成功')
      await load()
    },
  })
}
</script>

<template>
  <n-card>
    <div class="flex items-center justify-between mb-4">
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="搜索分组名称"
          clearable
          style="width: 220px"
          @keyup.enter="load"
        />
        <n-button @click="load">搜索</n-button>
      </n-space>
      <n-button type="primary" @click="openCreate">+ 新增分组</n-button>
    </div>
    <n-data-table :columns="columns" :data="list" :loading="loading" :bordered="false" />
  </n-card>

  <n-modal v-model:show="showForm" preset="card" :title="isEdit ? '编辑分组' : '新增分组'" style="width: 460px">
    <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="70">
      <n-form-item label="名称" path="name">
        <n-input v-model:value="form.name" placeholder="分组名称" />
      </n-form-item>
      <n-form-item label="备注" path="remark">
        <n-input v-model:value="form.remark" type="textarea" :autosize="{ minRows: 2 }" placeholder="可选" />
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
