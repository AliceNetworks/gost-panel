<template>
  <div class="port-forwards">
    <n-card>
      <template #header>
        <n-space justify="space-between" align="center">
          <span>端口转发</span>
          <n-button type="primary" @click="openCreateModal">
            添加转发规则
          </n-button>
        </n-space>
      </template>

      <!-- 骨架屏加载 -->
      <TableSkeleton v-if="loading && portForwards.length === 0" :rows="3" />

      <!-- 空状态 -->
      <EmptyState
        v-else-if="!loading && portForwards.length === 0"
        type="forwards"
        action-text="添加转发规则"
        @action="openCreateModal"
      />

      <!-- 数据表格 -->
      <n-data-table
        v-else
        :columns="columns"
        :data="portForwards"
        :loading="loading"
        :row-key="(row: any) => row.id"
      />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="editingForward ? '编辑转发规则' : '添加转发规则'" style="width: 600px;">
      <n-form :model="form" label-placement="left" label-width="120">
        <n-form-item label="规则名称">
          <n-input v-model:value="form.name" placeholder="例如: SSH-Forward" />
        </n-form-item>
        <n-form-item label="协议类型">
          <n-select v-model:value="form.protocol" :options="protocolOptions" />
        </n-form-item>

        <n-divider>源端配置</n-divider>
        <n-form-item label="监听地址">
          <n-input v-model:value="form.listen_host" placeholder="0.0.0.0 或留空" />
        </n-form-item>
        <n-form-item label="监听端口">
          <n-input-number v-model:value="form.listen_port" :min="1" :max="65535" style="width: 150px" />
        </n-form-item>

        <n-divider>目标端配置</n-divider>
        <n-form-item label="目标地址">
          <n-input v-model:value="form.target_host" placeholder="目标主机 IP 或域名" />
        </n-form-item>
        <n-form-item label="目标端口">
          <n-input-number v-model:value="form.target_port" :min="1" :max="65535" style="width: 150px" />
        </n-form-item>

        <n-divider>转发节点</n-divider>
        <n-form-item label="选择节点">
          <n-select v-model:value="form.node_id" :options="nodeOptions" filterable placeholder="选择执行转发的节点" clearable />
          <n-text depth="3" style="margin-top: 4px; font-size: 12px;">留空表示在本地执行转发</n-text>
        </n-form-item>

        <n-divider>其他选项</n-divider>
        <n-form-item label="启用">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>
        <n-form-item label="描述">
          <n-input v-model:value="form.description" type="textarea" placeholder="转发规则描述" :autosize="{ minRows: 2 }" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NButton, NSpace, NTag, NDropdown, useMessage, useDialog } from 'naive-ui'
import { getPortForwards, createPortForward, updatePortForward, deletePortForward, clonePortForward, getNodes } from '../api'
import EmptyState from '../components/EmptyState.vue'
import TableSkeleton from '../components/TableSkeleton.vue'
import { useKeyboard } from '../composables/useKeyboard'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const saving = ref(false)
const portForwards = ref<any[]>([])
const nodes = ref<any[]>([])
const showCreateModal = ref(false)
const editingForward = ref<any>(null)

const protocolOptions = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'RTCP (可靠 TCP)', value: 'rtcp' },
  { label: 'RUDP (可靠 UDP)', value: 'rudp' },
]

const defaultForm = () => ({
  name: '',
  protocol: 'tcp',
  listen_host: '0.0.0.0',
  listen_port: 8080,
  target_host: '',
  target_port: 80,
  node_id: null,
  enabled: true,
  description: '',
})

const form = ref(defaultForm())

const nodeOptions = ref<any[]>([])

const getProtocolLabel = (protocol: string) => {
  const opt = protocolOptions.find(o => o.value === protocol)
  return opt ? opt.label : protocol.toUpperCase()
}

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 150 },
  {
    title: '协议',
    key: 'protocol',
    width: 100,
    render: (row: any) => h(NTag, { type: 'info', size: 'small' }, () => getProtocolLabel(row.protocol)),
  },
  {
    title: '监听',
    key: 'listen',
    width: 200,
    render: (row: any) => `${row.listen_host || '0.0.0.0'}:${row.listen_port}`,
  },
  {
    title: '目标',
    key: 'target',
    width: 200,
    render: (row: any) => `${row.target_host}:${row.target_port}`,
  },
  {
    title: '节点',
    key: 'node_name',
    width: 120,
    render: (row: any) => row.node_name || '本地',
  },
  {
    title: '状态',
    key: 'enabled',
    width: 80,
    render: (row: any) =>
      h(NTag, { type: row.enabled ? 'success' : 'default', size: 'small' }, () => row.enabled ? '启用' : '禁用'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row: any) => {
      const dropdownOptions = [
        { label: '克隆', key: 'clone' },
        { type: 'divider', key: 'd1' },
        { label: '删除', key: 'delete' },
      ]
      const handleSelect = (key: string) => {
        switch (key) {
          case 'clone': handleClone(row); break
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

const loadPortForwards = async () => {
  loading.value = true
  try {
    const data: any = await getPortForwards()
    portForwards.value = data || []
  } catch (e) {
    message.error('加载端口转发失败')
  } finally {
    loading.value = false
  }
}

const loadNodes = async () => {
  try {
    const data: any = await getNodes()
    nodes.value = data || []
    nodeOptions.value = nodes.value.map((n: any) => ({
      label: `${n.name} (${n.host}:${n.port})`,
      value: n.id,
    }))
  } catch (e) {
    console.error('Failed to load nodes', e)
  }
}

const openCreateModal = () => {
  form.value = defaultForm()
  editingForward.value = null
  showCreateModal.value = true
}

const handleEdit = (row: any) => {
  editingForward.value = row
  form.value = { ...defaultForm(), ...row }
  showCreateModal.value = true
}

const handleSave = async () => {
  if (!form.value.name) {
    message.error('请输入规则名称')
    return
  }
  if (!form.value.target_host) {
    message.error('请输入目标地址')
    return
  }

  saving.value = true
  try {
    if (editingForward.value) {
      await updatePortForward(editingForward.value.id, form.value)
      message.success('转发规则已更新')
    } else {
      await createPortForward(form.value)
      message.success('转发规则已创建')
    }
    showCreateModal.value = false
    loadPortForwards()
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存转发规则失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = (row: any) => {
  dialog.warning({
    title: '删除转发规则',
    content: `确定要删除转发规则 "${row.name}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deletePortForward(row.id)
        message.success('转发规则已删除')
        loadPortForwards()
      } catch (e) {
        message.error('删除转发规则失败')
      }
    },
  })
}

const handleClone = async (row: any) => {
  try {
    await clonePortForward(row.id)
    message.success(`转发规则 "${row.name}" 已克隆`)
    loadPortForwards()
  } catch {
    message.error('克隆失败')
  }
}

onMounted(() => {
  loadPortForwards()
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
