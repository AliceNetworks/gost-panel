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

    <!-- Layout Customization Controls -->
    <n-space justify="space-between" align="center" class="mb-4">
      <span></span>
      <n-space>
        <n-button v-if="!editMode" quaternary size="small" @click="editMode = true">
          <template #icon>
            <n-icon><options-outline /></n-icon>
          </template>
          自定义布局
        </n-button>
        <template v-else>
          <n-button size="small" type="primary" @click="saveLayout">保存布局</n-button>
          <n-button size="small" @click="resetLayout">恢复默认</n-button>
          <n-button size="small" quaternary @click="editMode = false">取消</n-button>
        </template>
      </n-space>
    </n-space>

    <!-- Edit Mode: Card Visibility Toggle -->
    <n-card v-if="editMode" class="mb-4">
      <template #header>选择要显示的卡片</template>
      <n-space>
        <n-checkbox
          v-for="card in allCards"
          :key="card.id"
          :checked="visibleCardIds.includes(card.id)"
          @update:checked="(val: boolean) => toggleCard(card.id, val)"
        >
          {{ card.title }}
        </n-checkbox>
      </n-space>
      <n-text depth="3" style="font-size: 12px; margin-top: 8px; display: block;">
        提示: 拖拽卡片可以调整顺序
      </n-text>
    </n-card>

    <!-- Card Grid with Drag and Drop -->
    <div class="card-grid">
      <div
        v-for="cardId in visibleCardIds"
        :key="cardId"
        class="card-item"
        :class="{ 'edit-mode': editMode, 'dragging': draggedCardId === cardId, 'drag-over': dragOverCardId === cardId }"
        :draggable="editMode"
        @dragstart="onDragStart($event, cardId)"
        @dragover.prevent="onDragOver($event, cardId)"
        @drop="onDrop($event, cardId)"
        @dragend="onDragEnd"
      >
        <!-- User Plan Card -->
        <n-card v-if="cardId === 'user-plan' && userStore.user?.role !== 'admin'">
          <template #header>我的套餐</template>
          <n-space vertical>
            <n-descriptions :column="2" label-placement="left" size="small">
              <n-descriptions-item label="当前套餐">
                {{ userStore.user?.plan?.name || '无套餐' }}
              </n-descriptions-item>
              <n-descriptions-item label="到期时间">
                {{ userStore.user?.plan_expire_at ? formatDate(userStore.user.plan_expire_at) : '永久' }}
              </n-descriptions-item>
            </n-descriptions>
            <n-progress
              v-if="userStore.user?.plan?.traffic_quota"
              type="line"
              :percentage="trafficPercentage"
              :status="trafficPercentage > 90 ? 'error' : trafficPercentage > 70 ? 'warning' : 'success'"
            />
            <n-text v-if="userStore.user?.plan?.traffic_quota" depth="3">
              已使用 {{ formatBytes(userStore.user?.plan_traffic_used || 0) }} / {{ formatBytes(userStore.user?.plan?.traffic_quota || 0) }}
            </n-text>
          </n-space>
        </n-card>

        <!-- Stats Cards -->
        <n-grid v-if="cardId === 'stats'" :x-gap="16" :y-gap="16" :cols="4">
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
        <n-grid v-if="cardId === 'status-charts'" :x-gap="16" :y-gap="16" :cols="2">
          <n-grid-item>
            <n-card title="节点状态分布">
              <div :ref="(el: any) => onNodePieRef(el)" style="height: 200px"></div>
            </n-card>
          </n-grid-item>
          <n-grid-item>
            <n-card title="客户端状态分布">
              <div :ref="(el: any) => onClientPieRef(el)" style="height: 200px"></div>
            </n-card>
          </n-grid-item>
        </n-grid>

        <!-- Traffic Chart -->
        <n-card v-if="cardId === 'traffic-chart'" title="流量趋势">
          <template #header-extra>
            <n-select
              v-model:value="chartHours"
              :options="hoursOptions"
              style="width: 100px"
              size="small"
              @update:value="loadTrafficHistory"
            />
          </template>
          <div :ref="(el: any) => onTrafficChartRef(el)" style="height: 300px"></div>
        </n-card>

        <!-- Traffic Stats -->
        <n-card v-if="cardId === 'traffic-stats'" title="流量统计">
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
        <n-card v-if="cardId === 'nodes-status'" title="节点状态">
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, h, nextTick, computed } from 'vue'
import { NTag, NSwitch, NTooltip, NDivider, useMessage } from 'naive-ui'
import {
  ServerOutline,
  DesktopOutline,
  CheckmarkCircleOutline,
  RefreshOutline,
  NotificationsOutline,
  OptionsOutline,
} from '@vicons/ionicons5'
import * as echarts from 'echarts'
import { getStats, getNodes, getTrafficHistory } from '../api'
import { useBrowserNotification } from '../composables/useBrowserNotification'
import { useUserStore } from '../stores/user'
import { dashboardGuide, shouldShowGuide, markGuideComplete } from '../guides'

const message = useMessage()
const userStore = useUserStore()
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
let wsReconnectTimer: ReturnType<typeof setTimeout> | null = null
let resizeHandler: (() => void) | null = null

// Card customization state
const editMode = ref(false)
const draggedCardId = ref('')
const dragOverCardId = ref('')

// Card definitions
const allCards = [
  { id: 'user-plan', title: '我的套餐' },
  { id: 'stats', title: '系统概览' },
  { id: 'status-charts', title: '状态分布' },
  { id: 'traffic-chart', title: '流量趋势' },
  { id: 'traffic-stats', title: '流量统计' },
  { id: 'nodes-status', title: '节点状态' },
]

// Default layout
const getDefaultLayout = () => {
  if (userStore.user?.role !== 'admin') {
    return ['user-plan', 'stats', 'status-charts', 'traffic-chart', 'traffic-stats', 'nodes-status']
  }
  return ['stats', 'status-charts', 'traffic-chart', 'traffic-stats', 'nodes-status']
}

// Load layout from localStorage
const loadLayout = (): string[] => {
  try {
    const saved = localStorage.getItem('dashboard_layout')
    if (saved) {
      const layout = JSON.parse(saved)
      // Filter out invalid card IDs
      return layout.filter((id: string) => allCards.some(card => card.id === id))
    }
  } catch (e) {
    console.error('Failed to load dashboard layout', e)
  }
  return getDefaultLayout()
}

const visibleCardIds = ref<string[]>(loadLayout())

// Layout functions
const saveLayout = () => {
  try {
    localStorage.setItem('dashboard_layout', JSON.stringify(visibleCardIds.value))
    editMode.value = false
    message.success('布局已保存')
  } catch (e) {
    console.error('Failed to save layout', e)
    message.error('保存布局失败')
  }
}

const resetLayout = () => {
  visibleCardIds.value = getDefaultLayout()
  localStorage.removeItem('dashboard_layout')
  message.success('已恢复默认布局')
}

const toggleCard = (id: string, visible: boolean) => {
  if (visible) {
    if (!visibleCardIds.value.includes(id)) {
      visibleCardIds.value.push(id)
    }
  } else {
    visibleCardIds.value = visibleCardIds.value.filter(c => c !== id)
  }
}

// Drag and drop functions
const onDragStart = (e: DragEvent, id: string) => {
  draggedCardId.value = id
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', id)
  }
}

const onDragOver = (e: DragEvent, id: string) => {
  e.preventDefault()
  dragOverCardId.value = id
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = 'move'
  }
}

const onDrop = (e: DragEvent, targetId: string) => {
  e.preventDefault()
  dragOverCardId.value = ''

  if (draggedCardId.value && draggedCardId.value !== targetId) {
    const arr = [...visibleCardIds.value]
    const fromIdx = arr.indexOf(draggedCardId.value)
    const toIdx = arr.indexOf(targetId)

    if (fromIdx !== -1 && toIdx !== -1) {
      // Remove from old position
      arr.splice(fromIdx, 1)
      // Insert at new position
      arr.splice(toIdx, 0, draggedCardId.value)
      visibleCardIds.value = arr
    }
  }
}

const onDragEnd = () => {
  draggedCardId.value = ''
  dragOverCardId.value = ''
}

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

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

// Compute traffic percentage for user plan
const trafficPercentage = computed(() => {
  const quota = userStore.user?.plan?.traffic_quota || 0
  const used = userStore.user?.plan_traffic_used || 0
  if (quota === 0) return 0
  return Math.min(100, Math.round((used / quota) * 100))
})

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

// 函数 ref 回调 - 解决 v-for 内 template ref 不可靠的问题
const onTrafficChartRef = (el: HTMLElement | null) => {
  chartRef.value = el
  if (el) {
    if (chart) chart.dispose()
    chart = echarts.init(el)
    updateChart()
    if (!resizeHandler) {
      resizeHandler = () => {
        chart?.resize()
        nodePieChart?.resize()
        clientPieChart?.resize()
      }
      window.addEventListener('resize', resizeHandler)
    }
  }
}

const onNodePieRef = (el: HTMLElement | null) => {
  nodePieRef.value = el
  if (el) {
    if (nodePieChart) nodePieChart.dispose()
    nodePieChart = echarts.init(el)
    updatePieCharts()
  }
}

const onClientPieRef = (el: HTMLElement | null) => {
  clientPieRef.value = el
  if (el) {
    if (clientPieChart) clientPieChart.dispose()
    clientPieChart = echarts.init(el)
    updatePieCharts()
  }
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
let isUnmounted = false

const connectWebSocket = () => {
  if (isUnmounted) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`

  try {
    ws = new WebSocket(wsUrl)
  } catch {
    return
  }

  ws.onopen = () => {
    if (!isUnmounted) {
      wsConnected.value = true
    }
  }

  ws.onmessage = (event) => {
    if (isUnmounted) return
    try {
      const msg = JSON.parse(event.data)
      handleWSMessage(msg)
    } catch {
      // Ignore parse errors
    }
  }

  ws.onclose = () => {
    if (isUnmounted) return
    wsConnected.value = false
    // 清除之前的重连定时器
    if (wsReconnectTimer) {
      clearTimeout(wsReconnectTimer)
    }
    wsReconnectTimer = setTimeout(connectWebSocket, 5000)
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

onMounted(() => {
  // 不使用 await，让加载异步进行，不阻塞 UI
  loadAll()
  // 图表初始化由函数 ref 回调 (onTrafficChartRef/onNodePieRef/onClientPieRef) 自动处理
  startAutoRefresh()
  connectWebSocket()

  // Initialize notification state from localStorage
  const savedNotificationState = localStorage.getItem('notificationsEnabled')
  if (savedNotificationState === 'true' && checkPermission() === 'granted') {
    notificationsEnabled.value = true
  }

  // 首次访问时显示引导
  nextTick(() => {
    if (shouldShowGuide('dashboard')) {
      setTimeout(() => {
        dashboardGuide()
        markGuideComplete('dashboard')
      }, 500)
    }
  })
})

onUnmounted(() => {
  isUnmounted = true
  stopAutoRefresh()

  // 清除 resize 监听器
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }

  // 清除 WebSocket 重连定时器
  if (wsReconnectTimer) {
    clearTimeout(wsReconnectTimer)
    wsReconnectTimer = null
  }

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

/* Card Grid */
.card-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

.card-item {
  transition: all 0.2s ease;
}

.card-item.edit-mode {
  cursor: grab;
  border: 2px dashed transparent;
  border-radius: 8px;
  padding: 2px;
  margin: -2px;
}

.card-item.edit-mode:hover {
  border-color: rgba(59, 130, 246, 0.5);
  background-color: rgba(59, 130, 246, 0.05);
}

.card-item.edit-mode.dragging {
  opacity: 0.5;
  cursor: grabbing;
}

.card-item.edit-mode.drag-over {
  border-color: #3b82f6;
  background-color: rgba(59, 130, 246, 0.1);
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

:global(html:not(.dark)) .card-item.edit-mode:hover {
  border-color: rgba(59, 130, 246, 0.5);
  background-color: rgba(59, 130, 246, 0.05);
}

:global(html:not(.dark)) .card-item.edit-mode.drag-over {
  border-color: #3b82f6;
  background-color: rgba(59, 130, 246, 0.1);
}
</style>
