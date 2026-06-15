<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { useAuthStore } from '@/store/auth'
import { message } from '@/utils/discrete'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ 'update:show': [boolean] }>()

const auth = useAuthStore()
const formRef = ref<FormInst>()
const loading = ref(false)
const form = reactive({ old_password: '', new_password: '', confirm: '' })

const rules: FormRules = {
  old_password: { required: true, message: '请输入原密码', trigger: 'blur' },
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' },
  ],
  confirm: {
    required: true,
    trigger: 'blur',
    validator: () => form.confirm === form.new_password || new Error('两次输入的密码不一致'),
  },
}

// 关闭时清空表单
watch(
  () => props.show,
  (v) => {
    if (!v) Object.assign(form, { old_password: '', new_password: '', confirm: '' })
  },
)

async function submit() {
  await formRef.value?.validate()
  loading.value = true
  try {
    await auth.changePassword(form.old_password, form.new_password)
    message.success('密码修改成功')
    emit('update:show', false)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <n-modal
    :show="props.show"
    preset="card"
    title="修改密码"
    style="width: 420px"
    @update:show="emit('update:show', $event)"
  >
    <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
      <n-form-item label="原密码" path="old_password">
        <n-input v-model:value="form.old_password" type="password" show-password-on="click" placeholder="请输入原密码" />
      </n-form-item>
      <n-form-item label="新密码" path="new_password">
        <n-input v-model:value="form.new_password" type="password" show-password-on="click" placeholder="至少 6 位" />
      </n-form-item>
      <n-form-item label="确认密码" path="confirm">
        <n-input v-model:value="form.confirm" type="password" show-password-on="click" placeholder="再次输入新密码" @keyup.enter="submit" />
      </n-form-item>
    </n-form>
    <template #footer>
      <div class="flex justify-between">
        <n-button @click="emit('update:show', false)">取消</n-button>
        <n-button type="primary" :loading="loading" @click="submit">确定</n-button>
      </div>
    </template>
  </n-modal>
</template>
