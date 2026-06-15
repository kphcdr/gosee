<script setup lang="ts">
import type { TopItem } from '@/types/dashboard'
import { formatPercent } from '@/utils/format'

const props = defineProps<{
  items: TopItem[]
  title?: string
}>()

function barColor(v: number): string {
  if (v >= 90) return '#d03050'
  if (v >= 75) return '#f0a020'
  return '#18a058'
}
</script>

<template>
  <n-card :title="title">
    <n-space vertical :size="14">
      <div v-for="(item, i) in props.items" :key="item.server_id">
        <div class="flex items-center justify-between" style="font-size: 13px; margin-bottom: 4px">
          <span>
            <n-tag size="tiny" round :bordered="false" type="info" style="margin-right: 6px">{{ i + 1 }}</n-tag>
            {{ item.name }}
            <n-text depth="3" style="font-size: 11px; margin-left: 4px">{{ item.host }}</n-text>
          </span>
          <n-text strong>{{ formatPercent(item.value) }}</n-text>
        </div>
        <n-progress
          type="line"
          :percentage="Math.min(item.value, 100)"
          :height="8"
          :color="barColor(item.value)"
          rail-color="#efeff5"
          :show-indicator="false"
          :border-radius="4"
        />
      </div>
      <n-empty v-if="!props.items.length" description="暂无数据" />
    </n-space>
  </n-card>
</template>
