<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { menuOptions } from './menu'
import ChangePasswordModal from '@/components/ChangePasswordModal.vue'
import { APP_BUILD_TIME, APP_VERSION } from '@/constants/version'

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
      <div class="sider-content">
        <div class="logo">gosee 监控</div>
        <n-menu class="sider-menu" :options="menuOptions" :value="activeKey" :indent="18" @update:value="onSelect" />
        <div class="build-version" :title="`完整构建时间：${APP_BUILD_TIME}`">
          <span class="build-version-label">版本</span>
          <span class="build-version-value">{{ APP_VERSION }}</span>
        </div>
      </div>
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
.sider-content {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.logo {
  height: 56px;
  line-height: 56px;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #18a058;
  border-bottom: 1px solid #efeff5;
}

.sider-menu {
  flex: 1;
  min-height: 0;
}

.build-version {
  display: flex;
  align-items: baseline;
  gap: 6px;
  padding: 12px 16px 14px;
  border-top: 1px solid #efeff5;
  color: #8c8c92;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
}

.build-version-label {
  color: #a4a4aa;
}

.build-version-value {
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
}
</style>
