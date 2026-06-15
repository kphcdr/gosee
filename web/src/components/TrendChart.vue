<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    xData: string[]
    series: { name: string; data: number[] }[]
    unit?: string
    yMax?: number
    height?: number
  }>(),
  { height: 260 },
)

const el = ref<HTMLDivElement>()
let chart: echarts.ECharts | null = null
let ro: ResizeObserver | null = null

function render() {
  if (!chart) return
  const unit = props.unit ?? ''
  chart.setOption(
    {
      tooltip: {
        trigger: 'axis',
        valueFormatter: (v: unknown) => (v == null ? '-' : `${v}${unit}`),
      },
      legend: { data: props.series.map((s) => s.name), top: 0, textStyle: { fontSize: 11 } },
      grid: { left: 48, right: 16, top: 36, bottom: 28 },
      xAxis: {
        type: 'category',
        data: props.xData,
        boundaryGap: false,
        axisLabel: { fontSize: 10 },
      },
      yAxis: {
        type: 'value',
        max: props.yMax,
        axisLabel: { formatter: `{value}${unit}` },
      },
      series: props.series.map((s) => ({
        name: s.name,
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: s.data,
        areaStyle: { opacity: 0.1 },
      })),
    },
    true,
  )
}

onMounted(() => {
  if (!el.value) return
  chart = echarts.init(el.value)
  ro = new ResizeObserver(() => chart?.resize())
  ro.observe(el.value)
  render()
})

watch(() => [props.xData, props.series, props.yMax], render, { deep: true })

onBeforeUnmount(() => {
  ro?.disconnect()
  chart?.dispose()
  chart = null
})
</script>

<template>
  <div ref="el" :style="{ width: '100%', height: height + 'px' }" />
</template>
