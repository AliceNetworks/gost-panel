<template>
  <div class="operation-logs">
    <n-card>
      <template #header>
        <n-space justify="space-between" align="center">
          <span>操作日志</span>
          <n-space>
            <n-select
              v-model:value="filterAction"
              :options="actionOptions"
              placeholder="操作类型"
              clearable
              style="width: 120px"
              @update:value="handleFilter"
            />
            <n-select
              v-model:value="filterResource"
              :options="resourceOptions"
              placeholder="资源类型"
              clearable
              style="width: 120px"
              @update:value="handleFilter"
            />
            <n-button @click="loadLogs">
              <template #icon>
                <n-icon><refresh-outline /></n-icon>
              </template>
              刷新
            </n-button>
          </n-space>
        </n-space>
      </template>

      <!-- 骨架屏加载 -->
      <TableSkeleton v-if="loading && logs.length === 0" :rows="10" />

      <!-- 空状态 -->
      <EmptyState
        v-else-if="!loading && logs.length === 0"
        title="暂无操作日志"
        description="系统会自动记录用户操作"
      />

      <!-- 数据表格 -->
      <n-data-table
        v-else
        :columns="columns"
        :data="logs"
        :loading="loading"
        :row-key="(row: any) => row.id"
        :pagination="pagination"
        remote
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NTag, NIcon } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import { getOperationLogs } from '../api'
import EmptyState from '../components/EmptyState.vue'
import TableSkeleton from '../components/TableSkeleton.vue'

const loading = ref(false)
const logs = ref<any[]>([])
const filterAction = ref<string | null>(null)
const filterResource = ref<string | null>(null)

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [20, 50, 100],
})

const actionOptions = [
  { label: '登录', value: 'login' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '同步', value: 'sync' },
]

const resourceOptions = [
  { label: '用户', value: 'user' },
  { label: '节点', value: 'node' },
  { label: '客户端', value: 'client' },
  { label: '端口转发', value: 'port_forward' },
  { label: '节点组', value: 'node_group' },
  { label: '隧道', value: 'tunnel' },
  { label: '通知渠道', value: 'notify_channel' },
  { label: '告警规则', value: 'alert_rule' },
]

const formatTime = (time: string) => {
  if (!time || time.startsWith('0001')) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getActionTag = (action: string) => {
  const map: Record<string, { type: 'success' | 'info' | 'warning' | 'error' | 'default'; label: string }> = {
    login: { type: 'info', label: '登录' },
    create: { type: 'success', label: '创建' },
    update: { type: 'warning', label: '更新' },
    delete: { type: 'error', label: '删除' },
    sync: { type: 'info', label: '同步' },
  }
  return map[action] || { type: 'default', label: action }
}

const getResourceLabel = (resource: string) => {
  const map: Record<string, string> = {
    user: '用户',
    node: '节点',
    client: '客户端',
    port_forward: '端口转发',
    node_group: '节点组',
    tunnel: '隧道',
    notify_channel: '通知渠道',
    alert_rule: '告警规则',
    proxy_chain: '代理链',
  }
  return map[resource] || resource
}

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '时间',
    key: 'created_at',
    width: 170,
    render: (row: any) => formatTime(row.created_at)
  },
  {
    title: '用户',
    key: 'username',
    width: 100,
  },
  {
    title: '操作',
    key: 'action',
    width: 80,
    render: (row: any) => {
      const tag = getActionTag(row.action)
      return h(NTag, { type: tag.type, size: 'small' }, () => tag.label)
    }
  },
  {
    title: '资源',
    key: 'resource',
    width: 100,
    render: (row: any) => getResourceLabel(row.resource)
  },
  {
    title: '资源ID',
    key: 'resource_id',
    width: 80,
    render: (row: any) => row.resource_id || '-'
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row: any) => h(NTag, {
      type: row.status === 'success' ? 'success' : 'error',
      size: 'small'
    }, () => row.status === 'success' ? '成功' : '失败')
  },
  {
    title: 'IP',
    key: 'ip',
    width: 130,
  },
  {
    title: '详情',
    key: 'detail',
    ellipsis: {
      tooltip: true
    },
    render: (row: any) => {
      if (!row.detail) return '-'
      // 尝试解析 JSON
      try {
        const obj = JSON.parse(row.detail)
        return JSON.stringify(obj)
      } catch {
        return row.detail
      }
    }
  },
]

const loadLogs = async () => {
  loading.value = true
  try {
    const offset = (pagination.value.page - 1) * pagination.value.pageSize
    const data: any = await getOperationLogs({
      limit: pagination.value.pageSize,
      offset,
      action: filterAction.value || undefined,
      resource: filterResource.value || undefined,
    })
    logs.value = data.logs || []
    pagination.value.itemCount = data.total || 0
  } catch (e) {
    console.error('Failed to load operation logs', e)
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadLogs()
}

const handleFilter = () => {
  pagination.value.page = 1
  loadLogs()
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
</style>
