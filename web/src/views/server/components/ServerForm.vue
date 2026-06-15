<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { serverApi } from '@/api/server'
import { serverGroupApi } from '@/api/serverGroup'
import type { Server } from '@/types/server'
import type { ServerGroup } from '@/types/serverGroup'
import { AUTH_TYPE_OPTIONS, type AuthType } from '@/constants/server'
import { message } from '@/utils/discrete'

const props = defineProps<{ show: boolean; model: Server | null }>()
const emit = defineEmits<{ 'update:show': [boolean]; saved: [] }>()

const formRef = ref<FormInst>()
const saving = ref(false)
const groups = ref<ServerGroup[]>([])

const form = reactive({
  name: '',
  group_id: null as number | null,
  host: '',
  port: 22,
  username: 'root',
  auth_type: 'private_key' as AuthType,
  private_key: '',
  passphrase: '',
  password: '',
  remark: '',
  enabled: 1 as 0 | 1,
})

const isEdit = computed(() => !!props.model)
const groupOptions = computed(() => groups.value.map((g) => ({ label: g.name, value: g.id })))

const rules = computed<FormRules>(() => ({
  name: { required: true, message: '请输入服务器名称', trigger: 'blur' },
  host: { required: true, message: '请输入主机/IP', trigger: 'blur' },
  username: { required: true, message: '请输入登录用户', trigger: 'blur' },
  private_key: {
    trigger: 'blur',
    validator: () => {
      if (!isEdit.value && form.auth_type === 'private_key' && !form.private_key) return new Error('新建时必须填写私钥')
      return true
    },
  },
  password: {
    trigger: 'blur',
    validator: () => {
      if (!isEdit.value && form.auth_type === 'password' && !form.password) return new Error('新建时必须填写密码')
      return true
    },
  },
}))

watch(
  () => props.show,
  async (v) => {
    if (!v) return
    groups.value = await serverGroupApi.list().catch(() => [])
    if (props.model) {
      Object.assign(form, {
        name: props.model.name,
        group_id: props.model.group_id,
        host: props.model.host,
        port: props.model.port,
        username: props.model.username,
        auth_type: props.model.auth_type,
        remark: props.model.remark,
        enabled: props.model.enabled,
        private_key: '',
        passphrase: '',
        password: '',
      })
    } else {
      Object.assign(form, {
        name: '',
        group_id: null,
        host: '',
        port: 22,
        username: 'root',
        auth_type: 'private_key',
        remark: '',
        enabled: 1,
        private_key: '',
        passphrase: '',
        password: '',
      })
    }
  },
)

async function save() {
  await formRef.value?.validate()
  saving.value = true
  try {
    if (isEdit.value && props.model) {
      await serverApi.update(props.model.id, { ...form })
    } else {
      await serverApi.create({ ...form })
    }
    message.success('保存成功')
    emit('update:show', false)
    emit('saved')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    :title="isEdit ? '编辑服务器' : '新增服务器'"
    style="width: 580px"
    @update:show="emit('update:show', $event)"
  >
    <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="90">
      <n-form-item label="名称" path="name">
        <n-input v-model:value="form.name" placeholder="如 Web服务器-01" />
      </n-form-item>
      <n-form-item label="分组">
        <n-select v-model:value="form.group_id" :options="groupOptions" clearable placeholder="可选" />
      </n-form-item>
      <n-form-item label="主机/IP" path="host">
        <n-input v-model:value="form.host" placeholder="192.168.1.10 或域名" />
      </n-form-item>
      <n-form-item label="SSH 端口">
        <n-input-number
          :value="form.port"
          :min="1"
          :max="65535"
          style="width: 100%"
          @update:value="(v: number | null) => (form.port = v ?? 22)"
        />
      </n-form-item>
      <n-form-item label="登录用户" path="username">
        <n-input v-model:value="form.username" placeholder="root" />
      </n-form-item>
      <n-form-item label="认证方式">
        <n-radio-group v-model:value="form.auth_type">
          <n-radio-button v-for="o in AUTH_TYPE_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</n-radio-button>
        </n-radio-group>
      </n-form-item>
      <template v-if="form.auth_type === 'private_key'">
        <n-form-item label="私钥" path="private_key">
          <n-input
            v-model:value="form.private_key"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 8 }"
            :placeholder="isEdit ? '已配置，留空表示不修改；粘贴新私钥则覆盖' : '粘贴 SSH 私钥内容'"
          />
        </n-form-item>
        <n-form-item label="私钥口令">
          <n-input v-model:value="form.passphrase" type="password" show-password-on="click" placeholder="无私钥口令可留空" />
        </n-form-item>
      </template>
      <template v-else>
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            :placeholder="isEdit ? '已配置，留空表示不修改' : '请输入密码'"
          />
        </n-form-item>
      </template>
      <n-form-item label="备注">
        <n-input v-model:value="form.remark" type="textarea" :autosize="{ minRows: 2 }" placeholder="可选" />
      </n-form-item>
      <n-form-item label="启用">
        <n-switch :value="form.enabled === 1" @update:value="(v: boolean) => (form.enabled = v ? 1 : 0)" />
      </n-form-item>
    </n-form>
    <template #footer>
      <div class="flex justify-between">
        <n-button @click="emit('update:show', false)">取消</n-button>
        <n-button type="primary" :loading="saving" @click="save">保存</n-button>
      </div>
    </template>
  </n-modal>
</template>
