<template>
  <div class="dashboard">
    <!-- Refresh Controls -->
    <n-card class="mb-4">
      <n-space justify="space-between" align="center">
        <n-space align="center">
          <n-switch v-model:value="autoRefresh" @update:value="handleAutoRefreshChange">
            <template #checked>自动刷新</template>
            <template #unchecked>自动刷新</template>
          </n-switch>
          <n-select
            v-model:value="refreshInterval"
            :options="intervalOptions"
            :disabled="!autoRefresh"
            style="width: 120px"
            @update:value="handleIntervalChange"
          />
          <n-divider vertical />
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-switch v-model:value="notificationsEnabled" @update:value="handleNotificationToggle">
                <template #checked>
                  <n-icon><notifications-outline /></n-icon>
                </template>
                <template #unchecked>
                  <n-icon><notifications-outline /></n-icon>
                </template>
              </n-switch>
            </template>
            {{ notificationsEnabled ? '浏览器通知已开启' : '点击开启浏览器通知' }}
          </n-tooltip>
        </n-space>
        <n-space align="center">
          <n-tag v-if="wsConnected" type="success" size="small">
            实时
          </n-tag>
          <n-tag v-else type="default" size="small">
            离线
          </n-tag>
          <n-text depth="3" v-if="lastUpdate">
            更新时间: {{ formatTime(lastUpdate) }}
          </n-text>
          <n-button :loading="loading" @click="loadAll">
            <template #icon>
              <n-icon><refresh-outline /></n-icon>
            </template>
            刷新
          </n-button>
        </n-space>
      </n-space>
    </n-card>

    <!-- Stats Cards -->
    <n-grid :x-gap="16" :y-gap="16" :cols="4">
      <n-grid-item>
        <n-card>
          <n-statistic label="节点总数" :value="stats.total_nodes">
            <template #prefix>
              <n-icon><server-outline /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="在线节点" :value="stats.online_nodes">
            <template #prefix>
              <n-icon color="#18a058"><checkmark-circle-outline /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="客户端总数" :value="stats.total_clients">
            <template #prefix>
              <n-icon><desktop-outline /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card>
          <n-statistic label="在线客户端" :value="stats.online_clients">
            <template #prefix>
              <n-icon color="#18a058"><checkmark-circle-outline /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- Status Pie Charts -->
    <n-grid :x-gap="16" :y-gap="16" :cols="2" class="mt-4">
      <n-grid-item>
        <n-card title="节点状态分布">
          <div ref="nodePieRef" style="height: 200px"></div>
        </n-card>
      </n-grid-item>
      <n-grid-item>
        <n-card title="客户端状态分布">
          <div ref="clientPieRef" style="height: 200px"></div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- Traffic Chart -->
    <n-card title="流量趋势" class="mt-4">
      <template #header-extra>
        <n-select
          v-model:value="chartHours"
          :options="hoursOptions"
          style="width: 100px"
          size="small"
          @update:value="loadTrafficHistory"
        />
      </template>
      <div ref="chartRef" style="height: 300px"></div>
    </n-card>

    <!-- Traffic Stats -->
    <n-card title="流量统计" class="mt-4">
      <n-grid :x-gap="24" :cols="3">
        <n-grid-item>
          <div class="stat-row">
            <span class="label">入站流量</span>
            <span class="value">{{ formatBytes(stats.total_traffic_in) }}</span>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-row">
            <span class="label">出站流量</span>
            <span class="value">{{ formatBytes(stats.total_traffic_out) }}</span>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-row">
            <span class="label">当前连接</span>
            <span class="value highlight">{{ stats.total_connections }}</span>
          </div>
        </n-grid-item>
      </n-grid>
    </n-card>

    <!-- Nodes Status -->
    <n-card title="节点状态" class="mt-4">
      <n-data-table
        :columns="nodeColumns"
        :data="nodes"
        :loading="nodesLoading"
        :row-key="(row: any) => row.id"
        size="small"
        :max-height="300"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, h, nextTick } from 'vue'
import { NTag, NSwitch, NTooltip, NDivider } from 'naive-ui'
import {
  ServerOutline,
  DesktopOutline,
  CheckmarkCircleOutline,
  RefreshOutline,
  NotificationsOutline,
} from '@vicons/ionicons5'
import * as echarts from 'echarts'
import { getStats, getNodes, getTrafficHistory } from '../api'
import { useBrowserNotification } from '../composables/useBrowserNotification'

const { requestPermission, notifyNodeOffline, notifyNodeOnline, checkPermission } = useBrowserNotification()
const notificationsEnabled = ref(false)

const loading = ref(false)
const nodesLoading = ref(false)
const autoRefresh = ref(true)
const refreshInterval = ref(10000)
const lastUpdate = ref<Date | null>(null)
const chartHours = ref(1)
const chartRef = ref<HTMLElement | null>(null)
const nodePieRef = ref<HTMLElement | null>(null)
const clientPieRef = ref<HTMLElement | null>(null)
const wsConnected = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null
let chart: echarts.ECharts | null = null
let nodePieChart: echarts.ECharts | null = null
let clientPieChart: echarts.ECharts | null = null
let ws: WebSocket | null = null

const intervalOptions = [
  { label: '5s', value: 5000 },
  { label: '10s', value: 10000 },
  { label: '30s', value: 30000 },
  { label: '60s', value: 60000 },
]

const hoursOptions = [
  { label: '1 小时', value: 1 },
  { label: '6 小时', value: 6 },
  { label: '12 小时', value: 12 },
  { label: '24 小时', value: 24 },
]

const stats = ref({
  total_nodes: 0,
  online_nodes: 0,
  total_clients: 0,
  online_clients: 0,
  total_traffic_in: 0,
  total_traffic_out: 0,
  total_connections: 0,
})

const nodes = ref<any[]>([])
const trafficHistory = ref<any[]>([])

const nodeColumns = [
  { title: '名称', key: 'name', width: 120 },
  { title: '地址', key: 'host', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) =>
      h(NTag, { type: row.status === 'online' ? 'success' : 'default', size: 'small' }, () => row.status === 'online' ? '在线' : '离线'),
  },
  { title: '连接数', key: 'connections', width: 100 },
  {
    title: '入站流量',
    key: 'traffic_in',
    width: 120,
    render: (row: any) => formatBytes(row.traffic_in),
  },
  {
    title: '出站流量',
    key: 'traffic_out',
    width: 120,
    render: (row: any) => formatBytes(row.traffic_out),
  },
]

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (date: Date) => {
  return date.toLocaleTimeString()
}

const loadStats = async () => {
  loading.value = true
  try {
    const data: any = await getStats()
    stats.value = data
    lastUpdate.value = new Date()
    updatePieCharts()
  } catch (e) {
    console.error('Failed to load stats', e)
  } finally {
    loading.value = false
  }
}

const loadNodes = async () => {
  nodesLoading.value = true
  try {
    const data: any = await getNodes()
    nodes.value = data
  } catch (e) {
    console.error('Failed to load nodes', e)
  } finally {
    nodesLoading.value = false
  }
}

const loadTrafficHistory = async () => {
  try {
    const data: any = await getTrafficHistory(chartHours.value)
    trafficHistory.value = data || []
    updateChart()
  } catch (e) {
    console.error('Failed to load traffic history', e)
  }
}

const initChart = () => {
  if (!chartRef.value) return

  chart = echarts.init(chartRef.value)
  updateChart()

  // 监听窗口大小变化
  window.addEventListener('resize', () => {
    chart?.resize()
    nodePieChart?.resize()
    clientPieChart?.resize()
  })
}

const initPieCharts = () => {
  if (nodePieRef.value) {
    nodePieChart = echarts.init(nodePieRef.value)
  }
  if (clientPieRef.value) {
    clientPieChart = echarts.init(clientPieRef.value)
  }
  updatePieCharts()
}

const updatePieCharts = () => {
  const onlineNodes = stats.value.online_nodes
  const offlineNodes = stats.value.total_nodes - onlineNodes
  const onlineClients = stats.value.online_clients
  const offlineClients = stats.value.total_clients - onlineClients

  const pieOption = (online: number, offline: number, title: string): echarts.EChartsOption => ({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(15, 21, 53, 0.95)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      textStyle: { color: '#ffffff' },
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: 'rgba(255, 255, 255, 0.7)' }
    },
    series: [{
      name: title,
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 6,
        borderColor: 'rgba(0, 0, 0, 0.3)',
        borderWidth: 2
      },
      label: {
        show: true,
        position: 'center',
        formatter: () => `${online}/${online + offline}`,
        fontSize: 18,
        fontWeight: 'bold',
        color: '#fff'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: 'bold'
        }
      },
      data: [
        { value: online, name: '在线', itemStyle: { color: '#22c55e' } },
        { value: offline, name: '离线', itemStyle: { color: '#6b7280' } }
      ]
    }]
  })

  if (nodePieChart) {
    nodePieChart.setOption(pieOption(onlineNodes, offlineNodes, '节点状态'))
  }
  if (clientPieChart) {
    clientPieChart.setOption(pieOption(onlineClients, offlineClients, '客户端状态'))
  }
}

const updateChart = () => {
  if (!chart) return

  const times = trafficHistory.value.map((item: any) => {
    const date = new Date(item.time)
    return date.toLocaleTimeString()
  })

  const trafficIn = trafficHistory.value.map((item: any) => item.traffic_in / (1024 * 1024)) // Convert to MB
  const trafficOut = trafficHistory.value.map((item: any) => item.traffic_out / (1024 * 1024))
  const connections = trafficHistory.value.map((item: any) => item.connections)

  const option: echarts.EChartsOption = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(15, 21, 53, 0.95)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      borderWidth: 1,
      textStyle: {
        color: '#ffffff'
      },
      axisPointer: {
        type: 'cross',
        lineStyle: {
          color: 'rgba(255, 255, 255, 0.2)'
        }
      }
    },
    legend: {
      data: ['入站流量 (MB)', '出站流量 (MB)', '连接数'],
      textStyle: {
        color: 'rgba(255, 255, 255, 0.7)'
      },
      top: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '40px',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLine: {
        lineStyle: {
          color: 'rgba(255, 255, 255, 0.1)'
        }
      },
      axisLabel: {
        color: 'rgba(255, 255, 255, 0.5)'
      }
    },
    yAxis: [
      {
        type: 'value',
        name: '流量 (MB)',
        position: 'left',
        nameTextStyle: {
          color: 'rgba(255, 255, 255, 0.5)'
        },
        axisLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.1)'
          }
        },
        axisLabel: {
          color: 'rgba(255, 255, 255, 0.5)'
        },
        splitLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.05)'
          }
        }
      },
      {
        type: 'value',
        name: '连接数',
        position: 'right',
        nameTextStyle: {
          color: 'rgba(255, 255, 255, 0.5)'
        },
        axisLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.1)'
          }
        },
        axisLabel: {
          color: 'rgba(255, 255, 255, 0.5)'
        },
        splitLine: {
          show: false
        }
      }
    ],
    series: [
      {
        name: '入站流量 (MB)',
        type: 'line',
        smooth: true,
        data: trafficIn,
        itemStyle: { color: '#22c55e' },
        lineStyle: {
          width: 2,
          shadowColor: 'rgba(34, 197, 94, 0.3)',
          shadowBlur: 10
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(34, 197, 94, 0.3)' },
              { offset: 1, color: 'rgba(34, 197, 94, 0.02)' }
            ]
          }
        }
      },
      {
        name: '出站流量 (MB)',
        type: 'line',
        smooth: true,
        data: trafficOut,
        itemStyle: { color: '#3b82f6' },
        lineStyle: {
          width: 2,
          shadowColor: 'rgba(59, 130, 246, 0.3)',
          shadowBlur: 10
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
              { offset: 1, color: 'rgba(59, 130, 246, 0.02)' }
            ]
          }
        }
      },
      {
        name: '连接数',
        type: 'line',
        smooth: true,
        yAxisIndex: 1,
        data: connections,
        itemStyle: { color: '#f59e0b' },
        lineStyle: {
          width: 2,
          shadowColor: 'rgba(245, 158, 11, 0.3)',
          shadowBlur: 10
        }
      }
    ]
  }

  chart.setOption(option)
}

const loadAll = async () => {
  await Promise.all([loadStats(), loadNodes(), loadTrafficHistory()])
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  if (autoRefresh.value) {
    refreshTimer = setInterval(loadAll, refreshInterval.value)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const handleAutoRefreshChange = (value: boolean) => {
  if (value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

const handleIntervalChange = () => {
  if (autoRefresh.value) {
    startAutoRefresh()
  }
}

const handleNotificationToggle = async (value: boolean) => {
  if (value) {
    const granted = await requestPermission()
    if (!granted) {
      notificationsEnabled.value = false
    } else {
      localStorage.setItem('notificationsEnabled', 'true')
    }
  } else {
    localStorage.setItem('notificationsEnabled', 'false')
  }
}

// WebSocket connection
const connectWebSocket = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    wsConnected.value = true
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      handleWSMessage(msg)
    } catch {
      // Ignore parse errors
    }
  }

  ws.onclose = () => {
    wsConnected.value = false
    setTimeout(connectWebSocket, 5000)
  }

  ws.onerror = () => {
    ws?.close()
  }
}

const handleWSMessage = (msg: { type: string; data: any }) => {
  switch (msg.type) {
    case 'node_status':
      // Update node status in real-time
      const nodeData = msg.data
      const nodeIndex = nodes.value.findIndex((n: any) => n.id === nodeData.node_id)
      if (nodeIndex !== -1) {
        const oldStatus = nodes.value[nodeIndex].status
        const nodeName = nodes.value[nodeIndex].name
        nodes.value[nodeIndex] = {
          ...nodes.value[nodeIndex],
          status: nodeData.status,
          connections: nodeData.connections,
          traffic_in: nodeData.traffic_in,
          traffic_out: nodeData.traffic_out,
        }
        // Browser notification for status change
        if (notificationsEnabled.value && oldStatus !== nodeData.status) {
          if (nodeData.status === 'offline') {
            notifyNodeOffline(nodeData.node_id, nodeName)
          } else if (nodeData.status === 'online' && oldStatus === 'offline') {
            notifyNodeOnline(nodeData.node_id, nodeName)
          }
        }
      }
      lastUpdate.value = new Date()
      break

    case 'stats':
      // Update dashboard stats
      stats.value = msg.data
      lastUpdate.value = new Date()
      updatePieCharts()
      break
  }
}

onMounted(async () => {
  await loadAll()
  await nextTick()
  initChart()
  initPieCharts()
  startAutoRefresh()
  connectWebSocket()

  // Initialize notification state from localStorage
  const savedNotificationState = localStorage.getItem('notificationsEnabled')
  if (savedNotificationState === 'true' && checkPermission() === 'granted') {
    notificationsEnabled.value = true
  }
})

onUnmounted(() => {
  stopAutoRefresh()
  if (chart) {
    chart.dispose()
    chart = null
  }
  if (nodePieChart) {
    nodePieChart.dispose()
    nodePieChart = null
  }
  if (clientPieChart) {
    clientPieChart.dispose()
    clientPieChart = null
  }
  if (ws) {
    ws.close()
    ws = null
  }
})
</script>

<style scoped>
.dashboard {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-4 {
  margin-top: 16px;
}

.stat-row {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
}

.label {
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
  margin-bottom: 4px;
}

.value {
  font-weight: 600;
  font-size: 20px;
  color: rgba(255, 255, 255, 0.9);
}

.value.highlight {
  background: linear-gradient(135deg, #22c55e 0%, #06b6d4 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Stat Card Styles */
:deep(.n-statistic .n-statistic-value) {
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 700;
  font-size: 32px;
}

:deep(.n-statistic .n-statistic-label) {
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
  font-weight: 500;
}

/* Live indicator pulse */
.live-tag {
  animation: pulse-glow 2s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% {
    box-shadow: 0 0 5px rgba(34, 197, 94, 0.3);
  }
  50% {
    box-shadow: 0 0 15px rgba(34, 197, 94, 0.5), 0 0 25px rgba(34, 197, 94, 0.3);
  }
}

/* Card hover effect */
:deep(.n-card) {
  transition: all 0.25s ease;
}

:deep(.n-card:hover) {
  border-color: rgba(255, 255, 255, 0.15) !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), 0 0 20px rgba(59, 130, 246, 0.1);
}

/* Grid item animation */
:deep(.n-grid-item) {
  animation: fadeIn 0.3s ease-out;
  animation-fill-mode: both;
}

:deep(.n-grid-item:nth-child(1)) { animation-delay: 0.05s; }
:deep(.n-grid-item:nth-child(2)) { animation-delay: 0.1s; }
:deep(.n-grid-item:nth-child(3)) { animation-delay: 0.15s; }
:deep(.n-grid-item:nth-child(4)) { animation-delay: 0.2s; }

/* 亮色主题样式覆盖 */
:global(html:not(.dark)) .label {
  color: #718096;
}

:global(html:not(.dark)) .value {
  color: #2c3e50;
}

:global(html:not(.dark)) :deep(.n-statistic .n-statistic-label) {
  color: #718096;
}

:global(html:not(.dark)) :deep(.n-card:hover) {
  border-color: #e8e4db !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08), 0 0 10px rgba(79, 124, 255, 0.08);
}
</style>
