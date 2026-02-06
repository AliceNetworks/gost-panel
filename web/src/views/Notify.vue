<template>
  <div class="notify">
    <n-grid :x-gap="16" :y-gap="16" :cols="1">
      <!-- Notification Channels -->
      <n-grid-item>
        <n-card>
          <template #header>
            <n-space justify="space-between" align="center">
              <span>通知渠道</span>
              <n-button type="primary" @click="openCreateChannelModal">
                添加渠道
              </n-button>
            </n-space>
          </template>

          <!-- 骨架屏加载 -->
          <TableSkeleton v-if="channelsLoading && channels.length === 0" :rows="2" />

          <!-- 空状态 -->
          <EmptyState
            v-else-if="!channelsLoading && channels.length === 0"
            type="notify"
            action-text="添加渠道"
            @action="openCreateChannelModal"
          />

          <!-- 数据表格 -->
          <n-data-table
            v-else
            :columns="channelColumns"
            :data="channels"
            :loading="channelsLoading"
            :row-key="(row: any) => row.id"
          />
        </n-card>
      </n-grid-item>

      <!-- Alert Rules -->
      <n-grid-item>
        <n-card>
          <template #header>
            <n-space justify="space-between" align="center">
              <span>告警规则</span>
              <n-button type="primary" @click="openCreateRuleModal">
                添加规则
              </n-button>
            </n-space>
          </template>

          <!-- 骨架屏加载 -->
          <TableSkeleton v-if="rulesLoading && rules.length === 0" :rows="2" />

          <!-- 空状态 -->
          <EmptyState
            v-else-if="!rulesLoading && rules.length === 0"
            title="暂无告警规则"
            description="创建告警规则来监控节点状态"
            action-text="添加规则"
            @action="openCreateRuleModal"
          />

          <!-- 数据表格 -->
          <n-data-table
            v-else
            :columns="ruleColumns"
            :data="rules"
            :loading="rulesLoading"
            :row-key="(row: any) => row.id"
          />
        </n-card>
      </n-grid-item>

      <!-- Alert Logs -->
      <n-grid-item>
        <n-card title="告警日志">
          <n-data-table
            :columns="logColumns"
            :data="logs"
            :loading="logsLoading"
            :row-key="(row: any) => row.id"
            :pagination="logPagination"
            size="small"
            max-height="400"
          />
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- Channel Modal -->
    <n-modal v-model:show="showChannelModal" preset="dialog" :title="editingChannel ? '编辑通知渠道' : '添加通知渠道'" style="width: 600px;">
      <n-form :model="channelForm" label-placement="left" label-width="100">
        <n-form-item label="名称">
          <n-input v-model:value="channelForm.name" placeholder="例如: Telegram Bot" />
        </n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="channelForm.type" :options="channelTypeOptions" @update:value="handleChannelTypeChange" />
        </n-form-item>

        <!-- Telegram -->
        <template v-if="channelForm.type === 'telegram'">
          <n-form-item label="Bot Token">
            <n-input v-model:value="channelConfig.bot_token" placeholder="从 @BotFather 获取" />
          </n-form-item>
          <n-form-item label="Chat ID">
            <n-input v-model:value="channelConfig.chat_id" placeholder="你的 Chat ID" />
          </n-form-item>
        </template>

        <!-- Webhook -->
        <template v-if="channelForm.type === 'webhook'">
          <n-form-item label="Webhook URL">
            <n-input v-model:value="channelConfig.url" placeholder="https://your-webhook-url" />
          </n-form-item>
          <n-form-item label="HTTP 方法">
            <n-select v-model:value="channelConfig.method" :options="[{ label: 'POST', value: 'POST' }, { label: 'GET', value: 'GET' }]" />
          </n-form-item>
          <n-form-item label="Headers">
            <n-input v-model:value="channelConfig.headers" type="textarea" placeholder='{"Content-Type": "application/json"}' :autosize="{ minRows: 2 }" />
          </n-form-item>
        </template>

        <!-- Email -->
        <template v-if="channelForm.type === 'email'">
          <n-form-item label="SMTP 服务器">
            <n-input v-model:value="channelConfig.smtp_host" placeholder="smtp.gmail.com" />
          </n-form-item>
          <n-form-item label="SMTP 端口">
            <n-input-number v-model:value="channelConfig.smtp_port" :min="1" :max="65535" style="width: 150px" />
          </n-form-item>
          <n-form-item label="用户名">
            <n-input v-model:value="channelConfig.username" placeholder="your@email.com" />
          </n-form-item>
          <n-form-item label="密码">
            <n-input v-model:value="channelConfig.password" type="password" placeholder="SMTP 密码" />
          </n-form-item>
          <n-form-item label="发件人">
            <n-input v-model:value="channelConfig.from" placeholder="noreply@example.com" />
          </n-form-item>
          <n-form-item label="收件人">
            <n-input v-model:value="channelConfig.to" placeholder="admin@example.com (多个用逗号分隔)" />
          </n-form-item>
          <n-form-item label="使用 TLS">
            <n-switch v-model:value="channelConfig.use_tls" />
          </n-form-item>
        </template>

        <n-form-item label="启用">
          <n-switch v-model:value="channelForm.enabled" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showChannelModal = false">取消</n-button>
          <n-button v-if="editingChannel" type="info" :loading="testing" @click="handleTestChannel">测试</n-button>
          <n-button type="primary" :loading="saving" @click="handleSaveChannel">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Rule Modal -->
    <n-modal v-model:show="showRuleModal" preset="dialog" :title="editingRule ? '编辑告警规则' : '添加告警规则'" style="width: 600px;">
      <n-form :model="ruleForm" label-placement="left" label-width="120">
        <n-form-item label="规则名称">
          <n-input v-model:value="ruleForm.name" placeholder="例如: 节点离线告警" />
        </n-form-item>
        <n-form-item label="告警类型">
          <n-select v-model:value="ruleForm.alert_type" :options="alertTypeOptions" />
        </n-form-item>
        <n-form-item label="通知渠道">
          <n-select v-model:value="ruleForm.channel_ids" :options="channelOptions" multiple placeholder="选择一个或多个渠道" />
        </n-form-item>

        <n-divider>触发条件</n-divider>
        <template v-if="ruleForm.alert_type === 'node_offline'">
          <n-form-item label="离线时长">
            <n-space>
              <n-input-number v-model:value="ruleCondition.offline_duration" :min="1" style="width: 120px" />
              <span>分钟</span>
            </n-space>
          </n-form-item>
        </template>
        <template v-if="ruleForm.alert_type === 'quota_warning'">
          <n-form-item label="流量使用率">
            <n-space>
              <n-input-number v-model:value="ruleCondition.threshold" :min="1" :max="100" style="width: 120px" />
              <span>%</span>
            </n-space>
            <n-text depth="3" style="margin-top: 4px; font-size: 12px;">当流量使用达到此百分比时发送预警</n-text>
          </n-form-item>
        </template>
        <template v-if="ruleForm.alert_type === 'connection_limit'">
          <n-form-item label="连接数阈值">
            <n-input-number v-model:value="ruleCondition.max_connections" :min="1" style="width: 150px" />
          </n-form-item>
        </template>

        <n-divider>其他选项</n-divider>
        <n-form-item label="静默时间">
          <n-space>
            <n-input-number v-model:value="silenceDurationMin" :min="1" style="width: 120px" />
            <span>分钟</span>
          </n-space>
          <n-text depth="3" style="margin-top: 4px; font-size: 12px;">同一告警在静默时间内不会重复发送</n-text>
        </n-form-item>
        <n-form-item label="启用">
          <n-switch v-model:value="ruleForm.enabled" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showRuleModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSaveRule">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import { NButton, NSpace, NTag, useMessage, useDialog } from 'naive-ui'
import {
  getNotifyChannels,
  createNotifyChannel,
  updateNotifyChannel,
  deleteNotifyChannel,
  testNotifyChannel,
  getAlertRules,
  createAlertRule,
  updateAlertRule,
  deleteAlertRule,
  getAlertLogs,
} from '../api'
import EmptyState from '../components/EmptyState.vue'
import TableSkeleton from '../components/TableSkeleton.vue'

const message = useMessage()
const dialog = useDialog()

const channelsLoading = ref(false)
const rulesLoading = ref(false)
const logsLoading = ref(false)
const saving = ref(false)
const testing = ref(false)
const channels = ref<any[]>([])
const rules = ref<any[]>([])
const logs = ref<any[]>([])
const showChannelModal = ref(false)
const showRuleModal = ref(false)
const editingChannel = ref<any>(null)
const editingRule = ref<any>(null)

const channelTypeOptions = [
  { label: 'Telegram', value: 'telegram' },
  { label: 'Webhook', value: 'webhook' },
  { label: 'Email (SMTP)', value: 'email' },
]

const alertTypeOptions = [
  { label: '节点离线', value: 'node_offline' },
  { label: '节点恢复在线', value: 'node_online' },
  { label: '流量超限', value: 'quota_exceeded' },
  { label: '流量预警', value: 'quota_warning' },
  { label: '连接数告警', value: 'connection_limit' },
  { label: 'Agent 更新', value: 'agent_update' },
]

const defaultChannelForm = () => ({
  name: '',
  type: 'telegram',
  config: {},
  enabled: true,
})

const defaultRuleForm = () => ({
  name: '',
  alert_type: 'node_offline',
  channel_ids: [],
  condition: {},
  silence_duration: 300000,
  enabled: true,
})

const channelForm = ref(defaultChannelForm())
const ruleForm = ref(defaultRuleForm())
const channelConfig = ref<any>({})
const ruleCondition = ref<any>({})

const silenceDurationMin = computed({
  get: () => ruleForm.value.silence_duration / 60000,
  set: (val) => { ruleForm.value.silence_duration = val * 60000 }
})

const channelOptions = computed(() => {
  return channels.value
    .filter((c: any) => c.enabled)
    .map((c: any) => ({
      label: `${c.name} (${getChannelTypeLabel(c.type)})`,
      value: c.id,
    }))
})

const logPagination = ref({
  page: 1,
  pageSize: 20,
})

const getChannelTypeLabel = (type: string) => {
  const opt = channelTypeOptions.find(o => o.value === type)
  return opt ? opt.label : type
}

const getAlertTypeLabel = (type: string) => {
  const opt = alertTypeOptions.find(o => o.value === type)
  return opt ? opt.label : type
}

const formatTime = (time: string) => {
  if (!time) return '-'
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

const channelColumns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 150 },
  {
    title: '类型',
    key: 'type',
    width: 120,
    render: (row: any) => h(NTag, { type: 'info', size: 'small' }, () => getChannelTypeLabel(row.type)),
  },
  {
    title: '状态',
    key: 'enabled',
    width: 80,
    render: (row: any) =>
      h(NTag, { type: row.enabled ? 'success' : 'default', size: 'small' }, () => row.enabled ? '启用' : '禁用'),
  },
  {
    title: '最后测试',
    key: 'last_test_at',
    width: 150,
    render: (row: any) => formatTime(row.last_test_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => handleEditChannel(row) }, () => '编辑'),
        h(NButton, { size: 'small', type: 'info', onClick: () => handleTestChannelBtn(row) }, () => '测试'),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDeleteChannel(row) }, () => '删除'),
      ]),
  },
]

const ruleColumns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '规则名称', key: 'name', width: 180 },
  {
    title: '告警类型',
    key: 'alert_type',
    width: 150,
    render: (row: any) => h(NTag, { type: 'warning', size: 'small' }, () => getAlertTypeLabel(row.alert_type)),
  },
  {
    title: '通知渠道',
    key: 'channel_count',
    width: 100,
    render: (row: any) => `${row.channel_ids?.length || 0} 个`,
  },
  {
    title: '状态',
    key: 'enabled',
    width: 80,
    render: (row: any) =>
      h(NTag, { type: row.enabled ? 'success' : 'default', size: 'small' }, () => row.enabled ? '启用' : '禁用'),
  },
  {
    title: '触发次数',
    key: 'trigger_count',
    width: 100,
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => handleEditRule(row) }, () => '编辑'),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDeleteRule(row) }, () => '删除'),
      ]),
  },
]

const logColumns = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '告警类型',
    key: 'alert_type',
    width: 130,
    render: (row: any) => h(NTag, { type: 'warning', size: 'small' }, () => getAlertTypeLabel(row.alert_type)),
  },
  { title: '消息', key: 'message', ellipsis: { tooltip: true } },
  {
    title: '发送状态',
    key: 'sent',
    width: 100,
    render: (row: any) =>
      h(NTag, { type: row.sent ? 'success' : 'error', size: 'small' }, () => row.sent ? '已发送' : '失败'),
  },
  {
    title: '时间',
    key: 'created_at',
    width: 160,
    render: (row: any) => formatTime(row.created_at),
  },
]

const loadChannels = async () => {
  channelsLoading.value = true
  try {
    const data: any = await getNotifyChannels()
    channels.value = data || []
  } catch (e) {
    message.error('加载通知渠道失败')
  } finally {
    channelsLoading.value = false
  }
}

const loadRules = async () => {
  rulesLoading.value = true
  try {
    const data: any = await getAlertRules()
    rules.value = data || []
  } catch (e) {
    message.error('加载告警规则失败')
  } finally {
    rulesLoading.value = false
  }
}

const loadLogs = async () => {
  logsLoading.value = true
  try {
    const data: any = await getAlertLogs({ limit: 100 })
    logs.value = data || []
  } catch (e) {
    message.error('加载告警日志失败')
  } finally {
    logsLoading.value = false
  }
}

const openCreateChannelModal = () => {
  channelForm.value = defaultChannelForm()
  channelConfig.value = {}
  editingChannel.value = null
  showChannelModal.value = true
}

const handleEditChannel = (row: any) => {
  editingChannel.value = row
  channelForm.value = { ...defaultChannelForm(), ...row }
  channelConfig.value = { ...row.config }
  showChannelModal.value = true
}

const handleChannelTypeChange = () => {
  channelConfig.value = {}
  if (channelForm.value.type === 'telegram') {
    channelConfig.value = { bot_token: '', chat_id: '' }
  } else if (channelForm.value.type === 'webhook') {
    channelConfig.value = { url: '', method: 'POST', headers: '' }
  } else if (channelForm.value.type === 'email') {
    channelConfig.value = { smtp_host: '', smtp_port: 587, username: '', password: '', from: '', to: '', use_tls: true }
  }
}

const handleSaveChannel = async () => {
  if (!channelForm.value.name) {
    message.error('请输入名称')
    return
  }

  // 合并配置
  channelForm.value.config = channelConfig.value

  saving.value = true
  try {
    if (editingChannel.value) {
      await updateNotifyChannel(editingChannel.value.id, channelForm.value)
      message.success('通知渠道已更新')
    } else {
      await createNotifyChannel(channelForm.value)
      message.success('通知渠道已创建')
    }
    showChannelModal.value = false
    loadChannels()
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存通知渠道失败')
  } finally {
    saving.value = false
  }
}

const handleDeleteChannel = (row: any) => {
  dialog.warning({
    title: '删除通知渠道',
    content: `确定要删除通知渠道 "${row.name}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteNotifyChannel(row.id)
        message.success('通知渠道已删除')
        loadChannels()
      } catch (e) {
        message.error('删除通知渠道失败')
      }
    },
  })
}

const handleTestChannelBtn = async (row: any) => {
  testing.value = true
  try {
    await testNotifyChannel(row.id)
    message.success('测试通知已发送')
    loadChannels()
  } catch (e: any) {
    message.error(e.response?.data?.error || '发送测试通知失败')
  } finally {
    testing.value = false
  }
}

const handleTestChannel = async () => {
  if (!editingChannel.value) {
    message.error('请先保存渠道后再测试')
    return
  }
  await handleTestChannelBtn(editingChannel.value)
}

const openCreateRuleModal = () => {
  ruleForm.value = defaultRuleForm()
  ruleCondition.value = { offline_duration: 5 }
  editingRule.value = null
  showRuleModal.value = true
}

const handleEditRule = (row: any) => {
  editingRule.value = row
  ruleForm.value = { ...defaultRuleForm(), ...row }
  ruleCondition.value = { ...row.condition }
  showRuleModal.value = true
}

const handleSaveRule = async () => {
  if (!ruleForm.value.name) {
    message.error('请输入规则名称')
    return
  }
  if (!ruleForm.value.channel_ids || ruleForm.value.channel_ids.length === 0) {
    message.error('请选择至少一个通知渠道')
    return
  }

  // 合并条件
  ruleForm.value.condition = ruleCondition.value

  saving.value = true
  try {
    if (editingRule.value) {
      await updateAlertRule(editingRule.value.id, ruleForm.value)
      message.success('告警规则已更新')
    } else {
      await createAlertRule(ruleForm.value)
      message.success('告警规则已创建')
    }
    showRuleModal.value = false
    loadRules()
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存告警规则失败')
  } finally {
    saving.value = false
  }
}

const handleDeleteRule = (row: any) => {
  dialog.warning({
    title: '删除告警规则',
    content: `确定要删除告警规则 "${row.name}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteAlertRule(row.id)
        message.success('告警规则已删除')
        loadRules()
      } catch (e) {
        message.error('删除告警规则失败')
      }
    },
  })
}

onMounted(() => {
  loadChannels()
  loadRules()
  loadLogs()
})
</script>

<style scoped>
</style>
