<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FormInst, FormRules } from 'naive-ui'
import { useAuthStore } from '@/store/auth'
import { message } from '@/utils/discrete'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const formRef = ref<FormInst>()
const loading = ref(false)
const form = reactive({ username: 'admin', password: '' })

const rules: FormRules = {
  username: { required: true, message: '请输入用户名', trigger: 'blur' },
  password: { required: true, message: '请输入密码', trigger: 'blur' },
}

async function submit() {
  await formRef.value?.validate()
  loading.value = true
  try {
    await auth.login({ username: form.username, password: form.password })
    message.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.replace(redirect)
  } catch {
    // 错误已由 http 拦截器 message.error 统一提示
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrap">
    <n-card class="login-card" title="gosee 服务器监控" size="large" :bordered="false">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top" @keyup.enter="submit">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="用户名" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="密码" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="submit">登录</n-button>
      </n-form>
      <template #footer>
        <n-text depth="3" style="font-size: 12px">默认账号 admin / admin123</n-text>
      </template>
    </n-card>
  </div>
</template>

<style scoped>
.login-wrap {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #18a058 0%, #36ad6a 100%);
}
.login-card {
  width: 380px;
}
</style>
