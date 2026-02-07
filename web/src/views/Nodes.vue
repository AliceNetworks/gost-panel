<template>
  <div class="nodes">
    <n-card>
      <template #header>
        <n-space justify="space-between" align="center">
          <n-space align="center">
            <span>节点管理</span>
            <n-tag v-if="selectedRowKeys.length > 0" type="info" size="small">
              已选 {{ selectedRowKeys.length }} 项
            </n-tag>
          </n-space>
          <n-space>
            <!-- 批量操作按钮 -->
            <template v-if="selectedRowKeys.length > 0">
              <n-button type="warning" @click="handleBatchSync" :loading="batchLoading">
                批量同步
              </n-button>
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
            <n-button :loading="pingLoading" @click="handlePingAll">
              测试延迟
            </n-button>
            <n-button @click="openTemplateModal">
              快速配置
            </n-button>
            <n-button type="primary" @click="openCreateModal">
              添加节点
            </n-button>
          </n-space>
        </n-space>
      </template>

      <!-- 骨架屏加载 -->
      <TableSkeleton v-if="loading && nodes.length === 0" :rows="5" />

      <!-- 空状态 -->
      <EmptyState
        v-else-if="!loading && nodes.length === 0 && !searchText"
        type="nodes"
        action-text="添加节点"
        @action="openCreateModal"
      />

      <!-- 搜索无结果 -->
      <EmptyState
        v-else-if="!loading && nodes.length === 0 && searchText"
        type="search"
        :description="`未找到与 '${searchText}' 匹配的节点`"
      />

      <!-- 数据表格 -->
      <n-data-table
        v-else
        :columns="columns"
        :data="nodes"
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
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="editingNode ? '编辑节点' : '添加节点'" style="width: 700px;">
      <n-tabs type="line" animated>
        <!-- 基础配置 -->
        <n-tab-pane name="basic" tab="基础">
          <n-form :model="form" label-placement="left" label-width="120">
            <n-form-item label="名称">
              <n-input v-model:value="form.name" placeholder="例如: HK-1" />
            </n-form-item>
            <n-form-item label="地址">
              <n-input v-model:value="form.host" placeholder="例如: node.example.com" />
            </n-form-item>
            <n-form-item label="端口">
              <n-input-number v-model:value="form.port" :min="1" :max="65535" style="width: 150px" />
            </n-form-item>
            <n-divider style="margin: 12px 0;">
              <n-text depth="3" style="font-size: 12px;">GOST API 配置（用于远程同步配置）</n-text>
            </n-divider>
            <n-form-item label="API 端口">
              <n-input-number v-model:value="form.api_port" :min="1" :max="65535" style="width: 150px" />
            </n-form-item>
            <n-form-item label="API 用户">
              <n-input v-model:value="form.api_user" placeholder="GOST WebAPI 用户名" />
            </n-form-item>
            <n-form-item label="API 密码">
              <n-input-group>
                <n-input v-model:value="form.api_pass" type="password" placeholder="GOST WebAPI 密码" show-password-on="click" />
                <n-button @click="form.api_pass = generatePassword(16)">生成</n-button>
              </n-input-group>
            </n-form-item>
            <n-alert type="info" :bordered="false" style="font-size: 12px;">
              API 配置用于 Panel 直接调用 GOST 的 WebAPI 进行配置同步。如使用 Agent 模式，Agent 会自动配置。
            </n-alert>
          </n-form>
        </n-tab-pane>

        <!-- 协议配置 -->
        <n-tab-pane name="protocol" tab="协议">
          <n-form :model="form" label-placement="left" label-width="120">
            <n-form-item label="协议">
              <n-select v-model:value="form.protocol" :options="protocolOptions" style="width: 200px" />
            </n-form-item>
            <n-form-item label="传输层">
              <n-select v-model:value="form.transport" :options="transportOptions" style="width: 200px" />
            </n-form-item>

            <!-- SOCKS5/HTTP 认证 -->
            <template v-if="['socks5', 'http'].includes(form.protocol)">
              <n-form-item label="代理用户">
                <n-input v-model:value="form.proxy_user" placeholder="用户名" />
              </n-form-item>
              <n-form-item label="代理密码">
                <n-input v-model:value="form.proxy_pass" type="password" placeholder="密码" />
              </n-form-item>
            </template>

            <!-- Shadowsocks -->
            <template v-if="form.protocol === 'ss'">
              <n-form-item label="加密方式">
                <n-select v-model:value="form.ss_method" :options="ssMethodOptions" style="width: 250px" />
              </n-form-item>
              <n-form-item label="SS 密码">
                <n-input v-model:value="form.ss_password" type="password" placeholder="Shadowsocks 密码" />
              </n-form-item>
            </template>

            <!-- Trojan -->
            <template v-if="form.protocol === 'trojan'">
              <n-form-item label="密码">
                <n-input v-model:value="form.trojan_password" type="password" placeholder="Trojan 密码" />
              </n-form-item>
            </template>

            <!-- VMess -->
            <template v-if="form.protocol === 'vmess'">
              <n-form-item label="UUID">
                <n-input v-model:value="form.vmess_uuid" placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" />
              </n-form-item>
              <n-form-item label="Alter ID">
                <n-input-number v-model:value="form.vmess_alter_id" :min="0" :max="64" />
              </n-form-item>
            </template>
          </n-form>
        </n-tab-pane>

        <!-- 传输层配置 -->
        <n-tab-pane name="transport" tab="传输">
          <n-form :model="form" label-placement="left" label-width="120">
            <!-- TLS -->
            <template v-if="['tls', 'wss', 'h2', 'quic', 'grpc'].includes(form.transport)">
              <n-form-item label="启用 TLS">
                <n-switch v-model:value="form.tls_enabled" />
              </n-form-item>
              <n-form-item label="证书文件">
                <n-input v-model:value="form.tls_cert_file" placeholder="/path/to/cert.pem" />
              </n-form-item>
              <n-form-item label="密钥文件">
                <n-input v-model:value="form.tls_key_file" placeholder="/path/to/key.pem" />
              </n-form-item>
              <n-form-item label="SNI">
                <n-input v-model:value="form.tls_sni" placeholder="服务器名称" />
              </n-form-item>
            </template>

            <!-- WebSocket -->
            <template v-if="['ws', 'wss', 'mws', 'mwss'].includes(form.transport)">
              <n-form-item label="WS 路径">
                <n-input v-model:value="form.ws_path" placeholder="/ws" />
              </n-form-item>
              <n-form-item label="WS Host">
                <n-input v-model:value="form.ws_host" placeholder="example.com" />
              </n-form-item>
            </template>

            <!-- KCP 参数 -->
            <template v-if="form.transport === 'kcp'">
              <n-divider>KCP 参数</n-divider>
              <n-form-item label="MTU">
                <n-input-number v-model:value="kcpParams.mtu" :min="64" :max="9000" :default-value="1350" style="width: 150px" />
              </n-form-item>
              <n-form-item label="发送窗口">
                <n-input-number v-model:value="kcpParams.sndwnd" :min="1" :max="65535" :default-value="1024" style="width: 150px" />
              </n-form-item>
              <n-form-item label="接收窗口">
                <n-input-number v-model:value="kcpParams.rcvwnd" :min="1" :max="65535" :default-value="1024" style="width: 150px" />
              </n-form-item>
              <n-form-item label="数据分片">
                <n-input-number v-model:value="kcpParams.datashard" :min="0" :max="256" :default-value="10" style="width: 150px" />
              </n-form-item>
              <n-form-item label="校验分片">
                <n-input-number v-model:value="kcpParams.parityshard" :min="0" :max="256" :default-value="3" style="width: 150px" />
              </n-form-item>
            </template>

            <!-- TLS 高级选项 -->
            <template v-if="['tls', 'wss', 'h2', 'quic', 'grpc', 'mwss', 'http3', 'dtls'].includes(form.transport)">
              <n-divider>TLS 高级选项</n-divider>
              <n-form-item label="ALPN">
                <n-input v-model:value="form.tls_alpn" placeholder="h2,http/1.1 (逗号分隔)" />
              </n-form-item>
            </template>

            <!-- 限速 -->
            <n-divider>限速配置</n-divider>
            <n-form-item label="速度限制">
              <n-space>
                <n-input-number v-model:value="speedLimitMB" :min="0" :precision="2" style="width: 150px" />
                <span>MB/s (0 = 不限)</span>
              </n-space>
            </n-form-item>
            <n-form-item label="连接速率">
              <n-space>
                <n-input-number v-model:value="form.conn_rate_limit" :min="0" style="width: 150px" />
                <span>连接/秒 (0 = 不限)</span>
              </n-space>
            </n-form-item>

            <!-- DNS -->
            <n-divider>DNS 配置</n-divider>
            <n-form-item label="DNS 服务器">
              <n-input v-model:value="form.dns_server" placeholder="8.8.8.8:53 或 udp://1.1.1.1:53" />
            </n-form-item>

            <!-- 高级功能 -->
            <n-divider>高级功能</n-divider>
            <n-form-item label="PROXY Protocol">
              <n-select v-model:value="form.proxy_protocol" :options="proxyProtocolOptions" />
            </n-form-item>
            <n-form-item label="探测抵抗">
              <n-select v-model:value="form.probe_resist" :options="probeResistOptions" clearable placeholder="关闭" />
            </n-form-item>
            <n-form-item v-if="form.probe_resist" label="抵抗参数">
              <n-input v-model:value="form.probe_resist_value" :placeholder="probeResistPlaceholder" />
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 配额 -->
        <n-tab-pane name="quota" tab="配额">
          <n-form :model="form" label-placement="left" label-width="120">
            <n-form-item label="流量配额">
              <n-space>
                <n-input-number v-model:value="quotaGB" :min="0" :precision="2" style="width: 150px" />
                <span>GB (0 = 不限)</span>
              </n-space>
            </n-form-item>
            <n-form-item label="重置日">
              <n-space>
                <n-input-number v-model:value="form.quota_reset_day" :min="1" :max="28" style="width: 100px" />
                <span>每月第几天</span>
              </n-space>
            </n-form-item>
          </n-form>
        </n-tab-pane>
      </n-tabs>
      <template #action>
        <n-space>
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Config Modal -->
    <n-modal v-model:show="showConfigModal" preset="dialog" title="GOST 配置" style="width: 700px; max-width: 90vw;">
      <n-scrollbar x-scrollable style="max-height: 500px;">
        <n-code :code="configContent" language="yaml" word-wrap />
      </n-scrollbar>
      <template #action>
        <n-button @click="copyConfig">复制</n-button>
      </template>
    </n-modal>

    <!-- Install Script Modal -->
    <n-modal v-model:show="showScriptModal" preset="dialog" title="节点安装脚本" style="width: 750px; max-width: 90vw;">
      <n-tabs v-model:value="scriptOS" type="segment" style="margin-bottom: 16px;" @update:value="handleScriptOSChange">
        <n-tab-pane name="linux" tab="Linux / macOS" />
        <n-tab-pane name="windows" tab="Windows" />
      </n-tabs>
      <n-spin :show="scriptLoading">
        <n-alert type="info" style="margin-bottom: 16px;">
          {{ scriptOS === 'linux' ? '在目标服务器上运行以下命令：' : '在 PowerShell (管理员) 中运行：' }}
        </n-alert>
        <n-card size="small" style="margin-bottom: 16px; overflow: hidden;">
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

    <!-- Template Modal -->
    <n-modal v-model:show="showTemplateModal" preset="dialog" title="快速配置 - 选择模板" style="width: 800px;">
      <n-tabs type="segment" animated>
        <n-tab-pane v-for="cat in templateCategories" :key="cat.id" :name="cat.id" :tab="cat.name">
          <n-grid :x-gap="12" :y-gap="12" :cols="2">
            <n-grid-item v-for="tpl in getTemplatesByCategory(cat.id)" :key="tpl.id">
              <n-card
                hoverable
                :class="{ 'template-selected': selectedTemplate?.id === tpl.id }"
                @click="selectTemplate(tpl)"
                style="cursor: pointer;"
              >
                <template #header>
                  <n-space align="center">
                    <n-tag :type="getTagType(cat.id)" size="small">{{ tpl.defaults.protocol.toUpperCase() }}</n-tag>
                    <span>{{ tpl.name }}</span>
                  </n-space>
                </template>
                <n-text depth="3" style="font-size: 13px;">{{ tpl.description }}</n-text>
                <n-space style="margin-top: 8px;">
                  <n-tag size="tiny">{{ tpl.defaults.transport.toUpperCase() }}</n-tag>
                  <n-tag size="tiny" v-if="tpl.defaults.tls_enabled">TLS</n-tag>
                  <n-tag size="tiny">Port: {{ tpl.defaults.port }}</n-tag>
                </n-space>
              </n-card>
            </n-grid-item>
          </n-grid>
        </n-tab-pane>
      </n-tabs>
      <template #action>
        <n-space>
          <n-button @click="showTemplateModal = false">取消</n-button>
          <n-button type="primary" :disabled="!selectedTemplate" @click="applyTemplate">
            使用此模板
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Tag Management Modal -->
    <n-modal v-model:show="showTagModal" preset="dialog" title="管理节点标签" style="width: 500px;">
      <div v-if="editingNode" style="margin-bottom: 16px;">
        <n-text depth="3">节点: {{ editingNode.name }}</n-text>
      </div>

      <!-- 已有标签选择 -->
      <div style="margin-bottom: 16px;">
        <n-text strong style="display: block; margin-bottom: 8px;">选择标签</n-text>
        <n-space v-if="tags.length > 0">
          <n-tag
            v-for="tag in tags"
            :key="tag.id"
            :color="{ color: selectedTagIds.includes(tag.id) ? tag.color : 'transparent', textColor: selectedTagIds.includes(tag.id) ? '#fff' : tag.color, borderColor: tag.color }"
            :bordered="!selectedTagIds.includes(tag.id)"
            checkable
            :checked="selectedTagIds.includes(tag.id)"
            @update:checked="(checked: boolean) => { if (checked) selectedTagIds.push(tag.id); else selectedTagIds = selectedTagIds.filter(id => id !== tag.id) }"
            closable
            @close="handleDeleteTag(tag.id)"
            style="cursor: pointer;"
          >
            {{ tag.name }}
          </n-tag>
        </n-space>
        <n-text v-else depth="3">暂无标签，请先创建</n-text>
      </div>

      <!-- 创建新标签 -->
      <div style="margin-bottom: 16px;">
        <n-text strong style="display: block; margin-bottom: 8px;">创建新标签</n-text>
        <n-input-group>
          <n-input v-model:value="newTagName" placeholder="输入标签名称" @keyup.enter="handleCreateTag" />
          <n-button type="primary" @click="handleCreateTag">创建</n-button>
        </n-input-group>
      </div>

      <template #action>
        <n-space>
          <n-button @click="showTagModal = false">取消</n-button>
          <n-button type="primary" @click="handleSaveNodeTags">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Config Versions Modal -->
    <n-modal v-model:show="showVersionsModal" preset="dialog" :title="`配置历史: ${editingNode?.name}`" style="width: 800px;">
      <n-space vertical size="large">
        <n-space justify="space-between" align="center">
          <span>配置快照列表</span>
          <n-button type="primary" size="small" @click="openCreateVersionModal">
            创建快照
          </n-button>
        </n-space>

        <n-spin :show="versionsLoading">
          <n-list bordered v-if="configVersions.length > 0">
            <n-list-item v-for="version in configVersions" :key="version.id">
              <n-space vertical size="small" style="width: 100%">
                <n-space justify="space-between" align="center">
                  <n-space align="center">
                    <n-tag type="info" size="small">#{{ version.id }}</n-tag>
                    <n-text>{{ formatTime(version.created_at) }}</n-text>
                  </n-space>
                  <n-space>
                    <n-button size="small" @click="handleViewVersion(version)">查看</n-button>
                    <n-button size="small" type="primary" @click="handleRestoreVersion(version)">恢复</n-button>
                    <n-button size="small" type="error" @click="handleDeleteVersion(version)">删除</n-button>
                  </n-space>
                </n-space>
                <n-text depth="3" v-if="version.comment">{{ version.comment }}</n-text>
              </n-space>
            </n-list-item>
          </n-list>
          <n-empty v-else description="暂无配置快照" />
        </n-spin>
      </n-space>
      <template #action>
        <n-button @click="showVersionsModal = false">关闭</n-button>
      </template>
    </n-modal>

    <!-- Create Version Comment Modal -->
    <n-modal v-model:show="showVersionCommentModal" preset="dialog" title="创建配置快照" style="width: 500px;">
      <n-form label-placement="left" label-width="80">
        <n-form-item label="备注">
          <n-input v-model:value="versionComment" placeholder="可选：输入此快照的说明" type="textarea" :autosize="{ minRows: 3 }" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showVersionCommentModal = false">取消</n-button>
          <n-button type="primary" @click="handleCreateVersion">创建</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Version Config Modal -->
    <n-modal v-model:show="showVersionConfigModal" preset="dialog" title="配置内容" style="width: 700px; max-width: 90vw;">
      <n-scrollbar x-scrollable style="max-height: 500px;">
        <n-code :code="currentVersionConfig" language="yaml" word-wrap />
      </n-scrollbar>
      <template #action>
        <n-button @click="navigator.clipboard.writeText(currentVersionConfig); message.success('已复制到剪贴板')">复制</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import { NButton, NSpace, NTag, NProgress, NCollapse, NCollapseItem, NInputGroup, NText, NDivider, NTabs, NTabPane, NDropdown, NList, NListItem, NEmpty, NSpin, useMessage, useDialog } from 'naive-ui'
import { getNodesPaginated, createNode, updateNode, deleteNode, cloneNode, getNodeGostConfig, syncNodeConfig, getNodeProxyURI, getTemplates, getTemplateCategories, getNodeInstallScript, getTags, createTag, deleteTag, getNodeTags, setNodeTags, batchDeleteNodes, batchSyncNodes, pingNode, pingAllNodes, getConfigVersions, createConfigVersion, getConfigVersion, restoreConfigVersion, deleteConfigVersion } from '../api'
import EmptyState from '../components/EmptyState.vue'
import TableSkeleton from '../components/TableSkeleton.vue'
import { useKeyboard } from '../composables/useKeyboard'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const saving = ref(false)
const batchLoading = ref(false)
const nodes = ref<any[]>([])
const showCreateModal = ref(false)
const showConfigModal = ref(false)
const showTemplateModal = ref(false)
const showScriptModal = ref(false)
const showTagModal = ref(false)
const showVersionsModal = ref(false)
const scriptOS = ref('linux')
const scriptLoading = ref(false)
const currentScriptNodeId = ref<number | null>(null)
const configContent = ref('')
const installScript = ref('')
const oneLineCommand = ref('')
const editingNode = ref<any>(null)
const searchText = ref('')
const searchTimeout = ref<any>(null)

// 批量选择
const selectedRowKeys = ref<number[]>([])

// 延迟测试
const nodeLatencies = ref<Record<number, number>>({})
const pingLoading = ref(false)

// 标签相关
const tags = ref<any[]>([])
const selectedTagIds = ref<number[]>([])
const newTagName = ref('')
const tagColors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']

// 配置版本历史
const configVersions = ref<any[]>([])
const versionsLoading = ref(false)
const versionComment = ref('')
const showVersionCommentModal = ref(false)
const showVersionConfigModal = ref(false)
const currentVersionConfig = ref('')

// 模板相关
const templates = ref<any[]>([])
const templateCategories = ref<any[]>([])
const selectedTemplate = ref<any>(null)

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

// 协议选项
const protocolOptions = [
  { label: 'SOCKS5', value: 'socks5' },
  { label: 'HTTP/HTTPS', value: 'http' },
  { label: 'Shadowsocks', value: 'ss' },
  { label: 'Trojan', value: 'trojan' },
  { label: 'VMess', value: 'vmess' },
  { label: 'Relay (Port Forward)', value: 'relay' },
  { label: 'TCP Forward', value: 'tcp' },
  { label: 'UDP Forward', value: 'udp' },
  { label: 'SNI', value: 'sni' },
  { label: 'DNS', value: 'dns' },
  { label: 'SSH Tunnel (SSHD)', value: 'sshd' },
  { label: 'Redirect (透明代理)', value: 'redirect' },
  { label: 'TUN (全局代理)', value: 'tun' },
]

// 传输层选项
const transportOptions = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'TCP + UDP', value: 'tcp+udp' },
  { label: 'TLS', value: 'tls' },
  { label: 'mTLS', value: 'mtls' },
  { label: 'WebSocket (WS)', value: 'ws' },
  { label: 'WebSocket + TLS (WSS)', value: 'wss' },
  { label: 'Multiplex WS (mWS)', value: 'mws' },
  { label: 'Multiplex WS + TLS (mWSS)', value: 'mwss' },
  { label: 'HTTP/2 (H2)', value: 'h2' },
  { label: 'HTTP/2 Clear (H2C)', value: 'h2c' },
  { label: 'HTTP/3', value: 'http3' },
  { label: 'QUIC', value: 'quic' },
  { label: 'KCP', value: 'kcp' },
  { label: 'gRPC', value: 'grpc' },
  { label: 'PHT', value: 'pht' },
  { label: 'PHTS', value: 'phts' },
  { label: 'SSH', value: 'ssh' },
  { label: 'DTLS', value: 'dtls' },
  { label: 'Obfs-HTTP', value: 'ohttp' },
  { label: 'Obfs-TLS', value: 'otls' },
]

// SS 加密方法
const ssMethodOptions = [
  { label: 'AES-256-GCM (Recommended)', value: 'aes-256-gcm' },
  { label: 'AES-128-GCM', value: 'aes-128-gcm' },
  { label: 'ChaCha20-IETF-Poly1305', value: 'chacha20-ietf-poly1305' },
  { label: 'XCHACHA20-IETF-Poly1305', value: 'xchacha20-ietf-poly1305' },
  { label: '2022-BLAKE3-AES-256-GCM', value: '2022-blake3-aes-256-gcm' },
  { label: '2022-BLAKE3-CHACHA20-POLY1305', value: '2022-blake3-chacha20-poly1305' },
]

const proxyProtocolOptions = [
  { label: '关闭', value: 0 },
  { label: 'v1', value: 1 },
  { label: 'v2', value: 2 },
]

const probeResistOptions = [
  { label: '返回状态码 (code)', value: 'code' },
  { label: '代理到网站 (web)', value: 'web' },
  { label: '代理到主机 (host)', value: 'host' },
  { label: '返回文件 (file)', value: 'file' },
]

const probeResistPlaceholder = computed(() => {
  switch (form.value.probe_resist) {
    case 'code': return 'HTTP 状态码，例如: 404'
    case 'web': return '伪装网站 URL，例如: https://www.example.com'
    case 'host': return '转发主机地址，例如: example.com:443'
    case 'file': return '返回文件路径，例如: /var/www/index.html'
    default: return ''
  }
})

const defaultForm = () => ({
  name: '',
  host: '',
  port: 38567,
  api_port: 18080,
  api_user: 'admin',
  api_pass: '',
  proxy_user: '',
  proxy_pass: '',
  protocol: 'socks5',
  transport: 'tcp',
  ss_method: 'aes-256-gcm',
  ss_password: '',
  trojan_password: '',
  vmess_uuid: '',
  vmess_alter_id: 0,
  tls_enabled: false,
  tls_cert_file: '',
  tls_key_file: '',
  tls_sni: '',
  ws_path: '',
  ws_host: '',
  tls_alpn: '',
  speed_limit: 0,
  conn_rate_limit: 0,
  dns_server: '',
  proxy_protocol: 0,
  probe_resist: '',
  probe_resist_value: '',
  traffic_quota: 0,
  quota_reset_day: 1,
})

const form = ref(defaultForm())

const kcpParams = ref({
  mtu: 1350,
  sndwnd: 1024,
  rcvwnd: 1024,
  datashard: 10,
  parityshard: 3,
})

// 计算属性
const quotaGB = computed({
  get: () => form.value.traffic_quota / (1024 * 1024 * 1024),
  set: (val) => { form.value.traffic_quota = Math.round(val * 1024 * 1024 * 1024) }
})

const speedLimitMB = computed({
  get: () => form.value.speed_limit / (1024 * 1024),
  set: (val) => { form.value.speed_limit = Math.round(val * 1024 * 1024) }
})

// 格式化流量
const formatTraffic = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const columns = [
  { type: 'selection', width: 40 },
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name' },
  { title: '地址', key: 'host' },
  { title: '端口', key: 'port', width: 80 },
  {
    title: '协议',
    key: 'protocol',
    width: 100,
    render: (row: any) => h(NTag, { type: 'info', size: 'small' }, () => (row.protocol || 'socks5').toUpperCase()),
  },
  {
    title: '传输',
    key: 'transport',
    width: 80,
    render: (row: any) => h(NTag, { size: 'small' }, () => (row.transport || 'tcp').toUpperCase()),
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row: any) =>
      h(NTag, { type: row.status === 'online' ? 'success' : 'default', size: 'small' }, () => row.status === 'online' ? '在线' : '离线'),
  },
  {
    title: '延迟',
    key: 'latency',
    width: 80,
    render: (row: any) => {
      const latency = nodeLatencies.value[row.id]
      if (latency === undefined) {
        return h('span', { style: 'color: #999; font-size: 12px' }, '-')
      }
      if (latency === -1) {
        return h(NTag, { type: 'error', size: 'small' }, () => '超时')
      }
      const type = latency < 100 ? 'success' : latency < 300 ? 'warning' : 'error'
      return h(NTag, { type, size: 'small' }, () => `${latency}ms`)
    }
  },
  {
    title: '流量',
    key: 'traffic',
    width: 160,
    render: (row: any) => {
      if (row.traffic_quota > 0) {
        const percent = Math.min(100, (row.quota_used / row.traffic_quota) * 100)
        return h('div', { style: 'min-width: 140px' }, [
          h(NProgress, {
            type: 'line',
            percentage: percent,
            status: row.quota_exceeded ? 'error' : percent > 80 ? 'warning' : 'success',
            showIndicator: false,
            style: 'margin-bottom: 4px'
          }),
          h('span', { style: 'font-size: 11px; color: #666' },
            `${formatTraffic(row.quota_used)} / ${formatTraffic(row.traffic_quota)}`)
        ])
      }
      return h('span', { style: 'font-size: 12px' }, formatTraffic(row.traffic_in + row.traffic_out))
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render: (row: any) => {
      const dropdownOptions = [
        { label: '克隆节点', key: 'clone' },
        { label: '配置历史', key: 'versions' },
        { label: '安装脚本', key: 'install' },
        { label: '复制 URI', key: 'copy' },
        { label: '同步配置', key: 'sync' },
        { label: '查看配置', key: 'config' },
        { label: '管理标签', key: 'tags' },
        { label: '测试延迟', key: 'ping' },
        { type: 'divider', key: 'd1' },
        { label: '删除', key: 'delete' },
      ]
      const handleSelect = (key: string) => {
        switch (key) {
          case 'clone': handleCloneNode(row); break
          case 'versions': openVersionsModal(row); break
          case 'install': handleShowScript(row); break
          case 'copy': handleCopyURI(row); break
          case 'sync': handleSyncConfig(row); break
          case 'config': handleShowConfig(row); break
          case 'tags': openTagModal(row); break
          case 'ping': handlePingNode(row.id); break
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

const loadNodes = async () => {
  loading.value = true
  try {
    const data: any = await getNodesPaginated({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      search: searchText.value
    })
    nodes.value = data.items || []
    pagination.value.itemCount = data.total || 0
  } catch (e) {
    message.error('加载节点失败')
  } finally {
    loading.value = false
  }
}

// 测试所有节点延迟
const handlePingAll = async () => {
  pingLoading.value = true
  try {
    const data: any = await pingAllNodes()
    nodeLatencies.value = data.results || {}
  } catch (e) {
    message.error('延迟测试失败')
  } finally {
    pingLoading.value = false
  }
}

// 测试单个节点延迟
const handlePingNode = async (id: number) => {
  try {
    const data: any = await pingNode(id)
    nodeLatencies.value = { ...nodeLatencies.value, [id]: data.latency }
  } catch (e) {
    nodeLatencies.value = { ...nodeLatencies.value, [id]: -1 }
  }
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadNodes()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadNodes()
}

const handleSearch = () => {
  if (searchTimeout.value) clearTimeout(searchTimeout.value)
  searchTimeout.value = setTimeout(() => {
    pagination.value.page = 1
    loadNodes()
  }, 300)
}

const openCreateModal = () => {
  form.value = defaultForm()
  // 新建时自动生成 API 密码
  form.value.api_pass = generatePassword(16)
  editingNode.value = null
  showCreateModal.value = true
}

const handleEdit = (row: any) => {
  editingNode.value = row
  form.value = { ...defaultForm(), ...row }
  showCreateModal.value = true
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingNode.value) {
      await updateNode(editingNode.value.id, form.value)
      message.success('节点已更新')
    } else {
      await createNode(form.value)
      message.success('节点已创建')
    }
    showCreateModal.value = false
    loadNodes()
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存节点失败')
  } finally {
    saving.value = false
  }
}

const handleCloneNode = async (row: any) => {
  try {
    await cloneNode(row.id)
    message.success(`节点 "${row.name}" 已克隆`)
    loadNodes()
  } catch {
    message.error('克隆节点失败')
  }
}

const handleDelete = (row: any) => {
  dialog.warning({
    title: '删除节点',
    content: `确定要删除节点 "${row.name}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteNode(row.id)
        message.success('节点已删除')
        loadNodes()
      } catch (e) {
        message.error('删除节点失败')
      }
    },
  })
}

const handleShowConfig = async (row: any) => {
  try {
    const config: any = await getNodeGostConfig(row.id)
    configContent.value = typeof config === 'string' ? config : JSON.stringify(config, null, 2)
    showConfigModal.value = true
  } catch (e) {
    message.error('获取配置失败')
  }
}

const handleShowScript = async (row: any) => {
  try {
    currentScriptNodeId.value = row.id
    scriptOS.value = 'linux'
    const data: any = await getNodeInstallScript(row.id, 'linux')
    installScript.value = data.script || ''
    oneLineCommand.value = data.one_line_command || ''
    showScriptModal.value = true
  } catch (e) {
    message.error('获取安装脚本失败')
  }
}

const handleScriptOSChange = async (os: string) => {
  if (!currentScriptNodeId.value) return
  scriptLoading.value = true
  try {
    const data: any = await getNodeInstallScript(currentScriptNodeId.value, os)
    installScript.value = data.script || ''
    oneLineCommand.value = data.one_line_command || ''
  } catch (e) {
    message.error('获取安装脚本失败')
  } finally {
    scriptLoading.value = false
  }
}

const copyOneLineCommand = () => {
  navigator.clipboard.writeText(oneLineCommand.value)
  message.success('一键命令已复制到剪贴板')
}

const copyScript = () => {
  navigator.clipboard.writeText(installScript.value)
  message.success('已复制到剪贴板')
}

const handleSyncConfig = async (row: any) => {
  dialog.warning({
    title: '同步配置到节点',
    content: `这将把当前配置推送到节点 "${row.name}"。该节点的 GOST 服务必须正在运行并启用了 API。`,
    positiveText: '同步',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await syncNodeConfig(row.id)
        message.success('配置已成功同步到节点')
      } catch (e: any) {
        message.error(e.response?.data?.error || '同步配置失败')
      }
    },
  })
}

const handleCopyURI = async (row: any) => {
  try {
    const data: any = await getNodeProxyURI(row.id)
    navigator.clipboard.writeText(data.uri)
    message.success('代理 URI 已复制到剪贴板')
  } catch (e: any) {
    message.error(e.response?.data?.error || '获取代理 URI 失败')
  }
}

const copyConfig = () => {
  navigator.clipboard.writeText(configContent.value)
  message.success('已复制到剪贴板')
}

// 模板相关函数
const loadTemplates = async () => {
  try {
    const [tpls, cats]: any = await Promise.all([
      getTemplates(),
      getTemplateCategories()
    ])
    templates.value = tpls
    templateCategories.value = cats
  } catch (e) {
    console.error('Failed to load templates', e)
  }
}

const openTemplateModal = () => {
  selectedTemplate.value = null
  if (templates.value.length === 0) {
    loadTemplates()
  }
  showTemplateModal.value = true
}

const getTemplatesByCategory = (category: string) => {
  return templates.value.filter((t: any) => t.category === category)
}

const getTagType = (category: string) => {
  const types: Record<string, any> = {
    basic: 'default',
    secure: 'success',
    tunnel: 'warning',
    advanced: 'info'
  }
  return types[category] || 'default'
}

const selectTemplate = (tpl: any) => {
  selectedTemplate.value = tpl
}

const applyTemplate = () => {
  if (!selectedTemplate.value) return

  const tpl = selectedTemplate.value
  const defaults = tpl.defaults

  // 重置表单并应用模板默认值
  form.value = {
    ...defaultForm(),
    protocol: defaults.protocol,
    transport: defaults.transport,
    port: defaults.port,
    tls_enabled: defaults.tls_enabled || false,
    ws_path: defaults.ws_path || '',
    ss_method: defaults.ss_method || 'aes-256-gcm',
  }

  // 根据模板类型生成随机密码或 UUID
  if (defaults.require_auth) {
    form.value.proxy_user = 'user' + Math.random().toString(36).substring(2, 8)
    form.value.proxy_pass = generatePassword(16)
  }
  if (defaults.require_password) {
    if (defaults.protocol === 'ss') {
      form.value.ss_password = generatePassword(24)
    } else if (defaults.protocol === 'trojan') {
      form.value.trojan_password = generatePassword(24)
    }
  }
  if (defaults.require_uuid) {
    form.value.vmess_uuid = generateUUID()
  }

  // 生成 API 密码
  form.value.api_pass = generatePassword(16)

  showTemplateModal.value = false
  showCreateModal.value = true
  editingNode.value = null

  message.success(`模板 "${tpl.name}" 已应用。请填写其余字段。`)
}

const generatePassword = (length: number) => {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

const generateUUID = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

// ==================== 标签管理 ====================

const loadTags = async () => {
  try {
    tags.value = await getTags() as unknown as any[]
  } catch (e) {
    console.error('Failed to load tags', e)
  }
}

const openTagModal = async (node: any) => {
  editingNode.value = node
  try {
    const nodeTags = await getNodeTags(node.id) as unknown as any[]
    selectedTagIds.value = nodeTags.map((t: any) => t.id)
  } catch (e) {
    selectedTagIds.value = []
  }
  showTagModal.value = true
}

const handleCreateTag = async () => {
  if (!newTagName.value.trim()) return
  try {
    const randomColor = tagColors[Math.floor(Math.random() * tagColors.length)]
    await createTag({ name: newTagName.value.trim(), color: randomColor })
    newTagName.value = ''
    await loadTags()
    message.success('标签创建成功')
  } catch (e: any) {
    message.error(e.response?.data?.error || '创建标签失败')
  }
}

const handleDeleteTag = async (tagId: number) => {
  try {
    await deleteTag(tagId)
    await loadTags()
    // 从选中列表中移除
    selectedTagIds.value = selectedTagIds.value.filter(id => id !== tagId)
    message.success('标签已删除')
  } catch (e: any) {
    message.error(e.response?.data?.error || '删除标签失败')
  }
}

const handleSaveNodeTags = async () => {
  if (!editingNode.value) return
  try {
    await setNodeTags(editingNode.value.id, selectedTagIds.value)
    showTagModal.value = false
    await loadNodes()
    message.success('节点标签已更新')
  } catch (e: any) {
    message.error(e.response?.data?.error || '更新标签失败')
  }
}

// ==================== 批量操作 ====================

const handleCheckedRowKeysChange = (keys: (string | number)[]) => {
  selectedRowKeys.value = keys as number[]
}

const handleBatchDelete = () => {
  if (selectedRowKeys.value.length === 0) return

  dialog.warning({
    title: '批量删除确认',
    content: `确定要删除选中的 ${selectedRowKeys.value.length} 个节点吗？此操作不可恢复！`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      batchLoading.value = true
      try {
        const result: any = await batchDeleteNodes(selectedRowKeys.value)
        message.success(result.message || `成功删除 ${result.success} 个节点`)
        selectedRowKeys.value = []
        await loadNodes()
      } catch (e: any) {
        message.error(e.response?.data?.error || '批量删除失败')
      } finally {
        batchLoading.value = false
      }
    }
  })
}

const handleBatchSync = async () => {
  if (selectedRowKeys.value.length === 0) return

  batchLoading.value = true
  try {
    const result: any = await batchSyncNodes(selectedRowKeys.value)
    message.success(result.message || `成功同步 ${result.success} 个节点`)
    selectedRowKeys.value = []
  } catch (e: any) {
    message.error(e.response?.data?.error || '批量同步失败')
  } finally {
    batchLoading.value = false
  }
}

// ==================== 配置版本历史 ====================

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

const openVersionsModal = async (node: any) => {
  editingNode.value = node
  showVersionsModal.value = true
  await loadConfigVersions(node.id)
}

const loadConfigVersions = async (nodeId: number) => {
  versionsLoading.value = true
  try {
    configVersions.value = await getConfigVersions(nodeId) as unknown as any[]
  } catch (e) {
    message.error('加载配置历史失败')
  } finally {
    versionsLoading.value = false
  }
}

const openCreateVersionModal = () => {
  versionComment.value = ''
  showVersionCommentModal.value = true
}

const handleCreateVersion = async () => {
  if (!editingNode.value) return
  try {
    await createConfigVersion(editingNode.value.id, versionComment.value)
    message.success('配置快照已创建')
    showVersionCommentModal.value = false
    await loadConfigVersions(editingNode.value.id)
  } catch (e: any) {
    message.error(e.response?.data?.error || '创建快照失败')
  }
}

const handleViewVersion = async (version: any) => {
  try {
    const data: any = await getConfigVersion(version.id)
    currentVersionConfig.value = data.config || ''
    showVersionConfigModal.value = true
  } catch (e) {
    message.error('获取配置失败')
  }
}

const handleRestoreVersion = (version: any) => {
  dialog.warning({
    title: '恢复配置',
    content: `确定要恢复到版本 ${version.version} 吗？当前配置将被覆盖。`,
    positiveText: '恢复',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await restoreConfigVersion(version.id)
        message.success('配置已恢复')
        showVersionsModal.value = false
        loadNodes()
      } catch (e: any) {
        message.error(e.response?.data?.error || '恢复配置失败')
      }
    },
  })
}

const handleDeleteVersion = (version: any) => {
  dialog.warning({
    title: '删除快照',
    content: `确定要删除版本 ${version.version} 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteConfigVersion(version.id)
        message.success('快照已删除')
        if (editingNode.value) {
          await loadConfigVersions(editingNode.value.id)
        }
      } catch (e: any) {
        message.error(e.response?.data?.error || '删除快照失败')
      }
    },
  })
}

onMounted(() => {
  loadNodes()
  loadTags()
})

// Keyboard shortcuts
useKeyboard({
  onNew: openCreateModal,
  modalVisible: showCreateModal,
  onSave: handleSave,
})
</script>

<style scoped>
.template-selected {
  border: 2px solid #18a058;
  box-shadow: 0 0 8px rgba(24, 160, 88, 0.3);
}
</style>
