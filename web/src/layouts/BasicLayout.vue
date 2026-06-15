<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { menuOptions } from './menu'
import ChangePasswordModal from '@/components/ChangePasswordModal.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeKey = computed(() => route.path)
const showPwd = ref(false)

const userOptions = [
  { label: '修改密码', key: 'password' },
  { label: '退出登录', key: 'logout' },
]

function onSelect(key: string) {
  router.push(key)
}

onMounted(() => auth.fetchProfile())

function onUserMenu(key: string) {
  if (key === 'logout') {
    auth.logout()
    router.replace('/login')
  } else if (key === 'password') {
    showPwd.value = true
  }
}
</script>

<template>
  <n-layout has-sider style="height: 100vh">
    <n-layout-sider bordered :width="220" collapse-mode="width">
      <div class="logo">gosee 监控</div>
      <n-menu :options="menuOptions" :value="activeKey" :indent="18" @update:value="onSelect" />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered class="flex items-center justify-between px-6" style="height: 56px">
        <span style="font-size: 16px; font-weight: 500">{{ route.meta.title }}</span>
        <n-dropdown trigger="click" :options="userOptions" @select="onUserMenu">
          <n-button quaternary>{{ auth.user?.nickname || auth.user?.username || '用户' }}</n-button>
        </n-dropdown>
      </n-layout-header>
      <n-layout-content style="padding: 24px" :native-scrollbar="false">
        <router-view />
      </n-layout-content>
    </n-layout>
    <ChangePasswordModal v-model:show="showPwd" />
  </n-layout>
</template>

<style scoped>
.logo {
  height: 56px;
  line-height: 56px;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #18a058;
  border-bottom: 1px solid #efeff5;
}
</style>
