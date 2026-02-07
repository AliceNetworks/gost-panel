<template>
  <div class="clients">
    <n-card>
      <template #header>
        <n-space justify="space-between" align="center">
          <n-space align="center">
            <span>客户端管理（内网穿透）</span>
            <n-tag v-if="selectedRowKeys.length > 0" type="info" size="small">
              已选 {{ selectedRowKeys.length }} 项
            </n-tag>
          </n-space>
          <n-space>
            <!-- 批量操作按钮 -->
            <template v-if="selectedRowKeys.length > 0">
              <n-button type="error" @click="handleBatchDelete" :loading="batchLoading">
                批量删除
              </n-button>
              <n-divider vertical />
            </template>
            <n-input
              v-model:value="searchText"
              placeholder="搜索..."
              clearable
              style="width: 200px"
              @update:value="handleSearch"
            />
            <n-button type="primary" @click="openCreateModal">
              添加客户端
            </n-button>
          </n-space>
        </n-space>
      </template>

      <!-- 骨架屏加载 -->
      <TableSkeleton v-if="loading && clients.length === 0" :rows="5" />

      <!-- 空状态 -->
      <EmptyState
        v-else-if="!loading && clients.length === 0 && !searchText"
        type="clients"
        action-text="添加客户端"
        @action="openCreateModal"
      />

      <!-- 搜索无结果 -->
      <EmptyState
        v-else-if="!loading && clients.length === 0 && searchText"
        type="search"
        :description="`未找到与 '${searchText}' 匹配的客户端`"
      />

      <!-- 数据表格 -->
      <n-data-table
        v-else
        :columns="columns"
        :data="clients"
        :loading="loading"
        :row-key="(row: any) => row.id"
        :pagination="pagination"
        :checked-row-keys="selectedRowKeys"
        @update:checked-row-keys="handleCheckedRowKeysChange"
        remote
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="editingClient ? '编辑客户端' : '添加客户端'" style="width: 650px; max-width: 90vw;">
      <n-form :model="form" label-placement="left" label-width="100">
        <n-form-item label="名称" required>
          <n-input v-model:value="form.name" placeholder="例如: 家里电脑、公司内网" />
        </n-form-item>

        <n-form-item label="绑定节点" required>
          <n-select
            v-model:value="form.node_id"
            :options="nodeOptions"
            placeholder="选择入口节点（VPS）"
            filterable
          />
        </n-form-item>

        <n-divider>端口配置</n-divider>

        <n-grid :cols="2" :x-gap="12">
          <n-grid-item>
            <n-form-item label="本地端口">
              <n-input-number v-model:value="form.local_port" :min="1" :max="65535" style="width: 100%">
                <template #suffix>内网</template>
              </n-input-number>
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="远程端口">
              <n-input-number v-model:value="form.remote_port" :min="1" :max="65535" style="width: 100%">
                <template #suffix>VPS</template>
              </n-input-number>
            </n-form-item>
          </n-grid-item>
        </n-grid>

        <n-divider>代理认证（可选）</n-divider>

        <n-grid :cols="2" :x-gap="12">
          <n-grid-item>
            <n-form-item label="用户名">
              <n-input v-model:value="form.proxy_user" placeholder="代理用户名" />
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="密码">
              <n-input-group>
                <n-input v-model:value="form.proxy_pass" placeholder="代理密码" />
                <n-button @click="generateProxyPass">生成</n-button>
              </n-input-group>
            </n-form-item>
          </n-grid-item>
        </n-grid>

        <n-divider>流量配额（可选）</n-divider>

        <n-grid :cols="2" :x-gap="12">
          <n-grid-item>
            <n-form-item label="配额">
              <n-input-number v-model:value="form.traffic_quota_gb" :min="0" :precision="2" style="width: 100%">
                <template #suffix>GB</template>
              </n-input-number>
            </n-form-item>
          </n-grid-item>
          <n-grid-item>
            <n-form-item label="重置日">
              <n-input-number v-model:value="form.quota_reset_day" :min="1" :max="28" style="width: 100%">
                <template #suffix>日</template>
              </n-input-number>
            </n-form-item>
          </n-grid-item>
        </n-grid>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Install Script Modal -->
    <n-modal v-model:show="showScriptModal" preset="dialog" title="安装脚本" style="width: 750px; max-width: 90vw;">
      <n-tabs v-model:value="scriptOS" type="segment" style="margin-bottom: 16px;" @update:value="handleScriptOSChange">
        <n-tab-pane name="linux" tab="Linux / macOS" />
        <n-tab-pane name="windows" tab="Windows" />
      </n-tabs>
      <n-spin :show="scriptLoading">
        <n-alert type="info" style="margin-bottom: 16px;">
          {{ scriptOS === 'linux' ? '在内网设备上运行以下命令：' : '在 PowerShell (管理员) 中运行：' }}
        </n-alert>
        <n-card size="small" style="margin-bottom: 16px;">
          <n-scrollbar x-scrollable>
            <n-code :code="oneLineCommand" :language="scriptOS === 'linux' ? 'bash' : 'powershell'" word-wrap />
          </n-scrollbar>
        </n-card>
        <n-collapse>
          <n-collapse-item title="查看完整脚本" name="details">
            <n-scrollbar x-scrollable style="max-height: 300px;">
              <n-code :code="installScript" :language="scriptOS === 'linux' ? 'bash' : 'powershell'" word-wrap />
            </n-scrollbar>
          </n-collapse-item>
        </n-collapse>
      </n-spin>
      <template #action>
        <n-space>
          <n-button @click="copyOneLineCommand" :disabled="scriptLoading">复制一键命令</n-button>
          <n-button @click="copyScript" :disabled="scriptLoading">复制完整脚本</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- GOST Config Modal -->
    <n-modal v-model:show="showConfigModal" preset="dialog" title="GOST 配置" style="width: 700px; max-width: 90vw;">
      <n-alert type="info" style="margin-bottom: 16px;">
        这是客户端的 GOST 配置文件内容，保存到 /etc/gost/gost.yml
      </n-alert>
      <n-scrollbar x-scrollable style="max-height: 400px;">
        <n-code :code="configContent" language="yaml" word-wrap />
      </n-scrollbar>
      <template #action>
        <n-button @click="copyConfig">复制配置</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NButton, NSpace, NTag, NTabs, NTabPane, NDropdown, NDivider, useMessage, useDialog } from 'naive-ui'
import { getClientsPaginated, createClient, updateClient, deleteClient, getClientInstallScript, getClientGostConfig, getClientProxyURI, getNodes, batchDeleteClients, cloneClient } from '../api'
import EmptyState from '../components/EmptyState.vue'
import TableSkeleton from '../components/TableSkeleton.vue'
import { useKeyboard } from '../composables/useKeyboard'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const saving = ref(false)
const batchLoading = ref(false)
const clients = ref<any[]>([])
const nodes = ref<any[]>([])
const showCreateModal = ref(false)
const showScriptModal = ref(false)
const scriptOS = ref('linux')
const scriptLoading = ref(false)
const currentScriptClientId = ref<number | null>(null)
const showConfigModal = ref(false)
const installScript = ref('')
const oneLineCommand = ref('')
const configContent = ref('')
const editingClient = ref<any>(null)
const searchText = ref('')
const searchTimeout = ref<any>(null)

// 批量选择
const selectedRowKeys = ref<number[]>([])

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const defaultForm = () => ({
  name: '',
  node_id: null as number | null,
  local_port: 38777,
  remote_port: 38777,
  proxy_user: '',
  proxy_pass: '',
  traffic_quota_gb: 0,
  quota_reset_day: 1,
})

const form = ref(defaultForm())

const nodeOptions = ref<any[]>([])

const formatTraffic = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (time: string) => {
  if (!time || time.startsWith('0001')) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const columns = [
  { type: 'selection', width: 40 },
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 120 },
  {
    title: '节点',
    key: 'node',
    width: 120,
    render: (row: any) => row.node?.name || '-'
  },
  {
    title: '端口',
    key: 'ports',
    width: 120,
    render: (row: any) => `${row.local_port} → ${row.remote_port}`
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row: any) =>
      h(NTag, { type: row.status === 'online' ? 'success' : 'default', size: 'small' }, () => row.status === 'online' ? '在线' : '离线'),
  },
  {
    title: '流量',
    key: 'traffic',
    width: 120,
    render: (row: any) => `↑${formatTraffic(row.traffic_out)} ↓${formatTraffic(row.traffic_in)}`
  },
  {
    title: '最后心跳',
    key: 'last_seen',
    width: 150,
    render: (row: any) => formatTime(row.last_seen),
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row: any) => {
      const dropdownOptions = [
        { label: '克隆客户端', key: 'clone' },
        { label: '安装脚本', key: 'install' },
        { label: '复制 URI', key: 'copy' },
        { label: '查看配置', key: 'config' },
        { type: 'divider', key: 'd1' },
        { label: '删除', key: 'delete' },
      ]
      const handleSelect = (key: string) => {
        switch (key) {
          case 'clone': handleCloneClient(row); break
          case 'install': handleShowScript(row); break
          case 'copy': handleCopyURI(row); break
          case 'config': handleShowConfig(row); break
          case 'delete': handleDelete(row); break
        }
      }
      return h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => '编辑'),
        h(NDropdown, {
          options: dropdownOptions,
          onSelect: handleSelect,
          trigger: 'click'
        }, () => h(NButton, { size: 'small' }, () => '更多'))
      ])
    }
  },
]

const loadClients = async () => {
  loading.value = true
  try {
    const data: any = await getClientsPaginated({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      search: searchText.value
    })
    clients.value = data.items || []
    pagination.value.itemCount = data.total || 0
  } catch (e) {
    message.error('加载客户端失败')
  } finally {
    loading.value = false
  }
}

const loadNodes = async () => {
  try {
    const data: any = await getNodes()
    nodes.value = data
    nodeOptions.value = data.map((n: any) => ({
      label: `${n.name} (${n.host}:${n.port})`,
      value: n.id
    }))
  } catch (e) {
    console.error('Failed to load nodes', e)
  }
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadClients()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadClients()
}

const handleSearch = () => {
  if (searchTimeout.value) clearTimeout(searchTimeout.value)
  searchTimeout.value = setTimeout(() => {
    pagination.value.page = 1
    loadClients()
  }, 300)
}

const openCreateModal = () => {
  form.value = defaultForm()
  editingClient.value = null
  showCreateModal.value = true
}

const handleEdit = (row: any) => {
  editingClient.value = row
  form.value = {
    name: row.name,
    node_id: row.node_id,
    local_port: row.local_port,
    remote_port: row.remote_port,
    proxy_user: row.proxy_user || '',
    proxy_pass: row.proxy_pass || '',
    traffic_quota_gb: row.traffic_quota ? row.traffic_quota / (1024 * 1024 * 1024) : 0,
    quota_reset_day: row.quota_reset_day || 1,
  }
  showCreateModal.value = true
}

const handleSave = async () => {
  if (!form.value.name) {
    message.warning('请输入名称')
    return
  }
  if (!form.value.node_id) {
    message.warning('请选择节点')
    return
  }

  saving.value = true
  try {
    const payload = {
      ...form.value,
      traffic_quota: Math.round(form.value.traffic_quota_gb * 1024 * 1024 * 1024),
    }
    delete (payload as any).traffic_quota_gb

    if (editingClient.value) {
      await updateClient(editingClient.value.id, payload)
      message.success('客户端已更新')
    } else {
      await createClient(payload)
      message.success('客户端已创建')
    }
    showCreateModal.value = false
    loadClients()
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存客户端失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = (row: any) => {
  dialog.warning({
    title: '删除客户端',
    content: `确定要删除客户端 "${row.name}" 吗？远程设备上的 Agent 将在下次心跳时自动卸载。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteClient(row.id)
        message.success('客户端已删除')
        loadClients()
      } catch (e) {
        message.error('删除客户端失败')
      }
    },
  })
}

const handleCloneClient = async (row: any) => {
  try {
    await cloneClient(row.id)
    message.success(`客户端 "${row.name}" 已克隆`)
    loadClients()
  } catch {
    message.error('克隆客户端失败')
  }
}

const handleShowScript = async (row: any) => {
  try {
    currentScriptClientId.value = row.id
    scriptOS.value = 'linux'
    const data: any = await getClientInstallScript(row.id, 'linux')
    installScript.value = data.script || ''
    oneLineCommand.value = data.one_line_command || ''
    showScriptModal.value = true
  } catch (e) {
    message.error('获取安装脚本失败')
  }
}

const handleScriptOSChange = async (os: string) => {
  if (!currentScriptClientId.value) return
  scriptLoading.value = true
  try {
    const data: any = await getClientInstallScript(currentScriptClientId.value, os)
    installScript.value = data.script || ''
    oneLineCommand.value = data.one_line_command || ''
  } catch (e) {
    message.error('获取安装脚本失败')
  } finally {
    scriptLoading.value = false
  }
}

const handleShowConfig = async (row: any) => {
  try {
    const config: any = await getClientGostConfig(row.id)
    configContent.value = typeof config === 'string' ? config : JSON.stringify(config, null, 2)
    showConfigModal.value = true
  } catch (e) {
    message.error('获取配置失败')
  }
}

const copyOneLineCommand = () => {
  navigator.clipboard.writeText(oneLineCommand.value)
  message.success('已复制到剪贴板')
}

const copyScript = () => {
  navigator.clipboard.writeText(installScript.value)
  message.success('已复制到剪贴板')
}

const copyConfig = () => {
  navigator.clipboard.writeText(configContent.value)
  message.success('已复制到剪贴板')
}

const handleCopyURI = async (row: any) => {
  try {
    const data: any = await getClientProxyURI(row.id)
    navigator.clipboard.writeText(data.uri || '')
    message.success('代理 URI 已复制到剪贴板')
  } catch (e) {
    message.error('获取代理 URI 失败')
  }
}

const generateProxyPass = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  let result = ''
  for (let i = 0; i < 16; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  form.value.proxy_pass = result
}

// ==================== 批量操作 ====================

const handleCheckedRowKeysChange = (keys: (string | number)[]) => {
  selectedRowKeys.value = keys as number[]
}

const handleBatchDelete = () => {
  if (selectedRowKeys.value.length === 0) return

  dialog.warning({
    title: '批量删除确认',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个客户端吗？此操作不可恢复！`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      batchLoading.value = true
      try {
        const result: any = await batchDeleteClients(selectedRowKeys.value)
        message.success(result.message || `成功删除 ${result.success} 个客户端`)
        selectedRowKeys.value = []
        await loadClients()
      } catch (e: any) {
        message.error(e.response?.data?.error || '批量删除失败')
      } finally {
        batchLoading.value = false
      }
    }
  })
}

onMounted(() => {
  loadClients()
  loadNodes()
})

// Keyboard shortcuts
useKeyboard({
  onNew: openCreateModal,
  modalVisible: showCreateModal,
  onSave: handleSave,
})
</script>

<style scoped>
</style>
