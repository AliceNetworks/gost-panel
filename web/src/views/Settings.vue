<template>
  <div class="settings">
    <n-card>
      <template #header>
        <n-space justify="space-between" align="center">
          <span>网站设置</span>
          <n-button type="primary" :loading="saving" @click="handleSave">
            保存设置
          </n-button>
        </n-space>
      </template>

      <n-form :model="form" label-placement="left" label-width="140">
        <n-divider>基本信息</n-divider>

        <n-form-item label="网站名称">
          <n-input v-model:value="form.site_name" placeholder="GOST Panel" />
        </n-form-item>

        <n-form-item label="网站描述">
          <n-input v-model:value="form.site_description" placeholder="GOST 代理管理面板" />
        </n-form-item>

        <n-form-item label="网站 URL">
          <n-input v-model:value="form.site_url" placeholder="https://your-domain.com (用于邮件链接)" />
        </n-form-item>

        <n-divider>用户注册</n-divider>

        <n-form-item label="开放注册">
          <n-space vertical>
            <n-switch v-model:value="form.registration_enabled" />
            <n-text depth="3" style="font-size: 12px;">
              启用后，新用户可以自行注册账户
            </n-text>
          </n-space>
        </n-form-item>

        <n-form-item label="邮箱验证">
          <n-space vertical>
            <n-switch v-model:value="form.email_verification_required" :disabled="!form.registration_enabled" />
            <n-text depth="3" style="font-size: 12px;">
              启用后，用户必须验证邮箱才能登录（需先配置 SMTP）
            </n-text>
          </n-space>
        </n-form-item>

        <n-form-item label="默认角色">
          <n-select
            v-model:value="form.default_role"
            :options="roleOptions"
            :disabled="!form.registration_enabled"
            style="width: 200px;"
          />
        </n-form-item>

        <n-divider>图标配置</n-divider>

        <n-form-item label="Favicon URL">
          <n-input v-model:value="form.favicon_url" placeholder="/vite.svg 或 https://example.com/favicon.ico">
            <template #prefix>
              <img v-if="form.favicon_url" :src="form.favicon_url" style="width: 16px; height: 16px;" @error="onFaviconError" />
            </template>
          </n-input>
        </n-form-item>

        <n-form-item label="Logo URL">
          <n-input v-model:value="form.logo_url" placeholder="留空使用默认，或输入图片URL" />
        </n-form-item>

        <n-divider>自定义内容</n-divider>

        <n-form-item label="页脚文字">
          <n-input v-model:value="form.footer_text" placeholder="可选，显示在页面底部" />
        </n-form-item>

        <n-form-item label="自定义 CSS">
          <n-input
            v-model:value="form.custom_css"
            type="textarea"
            placeholder="可选，自定义样式"
            :rows="4"
          />
        </n-form-item>

        <n-divider>Agent 更新</n-divider>

        <n-form-item label="自动更新">
          <n-space vertical>
            <n-switch v-model:value="form.agent_auto_update" />
            <n-text depth="3" style="font-size: 12px;">
              启用后，Agent 会在心跳时检查更新并提示
            </n-text>
          </n-space>
        </n-form-item>

        <n-form-item label="强制更新">
          <n-space vertical>
            <n-switch v-model:value="form.agent_force_update" :disabled="!form.agent_auto_update" />
            <n-text depth="3" style="font-size: 12px;">
              启用后，所有 Agent 将在下次心跳时自动更新（更新后自动关闭）
            </n-text>
          </n-space>
        </n-form-item>

        <n-form-item label="当前 Agent 版本">
          <n-text>v1.3.0</n-text>
        </n-form-item>
      </n-form>

      <n-divider>预览</n-divider>
      <n-space vertical>
        <n-text>浏览器标签预览:</n-text>
        <n-space align="center">
          <img :src="form.favicon_url || '/vite.svg'" style="width: 20px; height: 20px;" />
          <n-text strong>{{ form.site_name || 'GOST Panel' }}</n-text>
        </n-space>
      </n-space>
    </n-card>

    <!-- 数据导出/导入 -->
    <n-card style="margin-top: 16px;">
      <template #header>
        <span>数据导出/导入</span>
      </template>
      <n-space vertical>
        <n-text depth="3">导出或导入节点和客户端配置，可用于备份或迁移。</n-text>
        <n-space>
          <n-select
            v-model:value="exportType"
            :options="exportTypeOptions"
            style="width: 140px"
          />
          <n-button :loading="exporting" @click="handleExport('json')">
            导出 JSON
          </n-button>
          <n-button :loading="exporting" @click="handleExport('yaml')">
            导出 YAML
          </n-button>
          <n-divider vertical />
          <n-upload
            :show-file-list="false"
            accept=".json,.yaml,.yml"
            :custom-request="handleImport"
          >
            <n-button :loading="importing" type="primary">
              导入配置
            </n-button>
          </n-upload>
        </n-space>
        <n-text depth="3" style="font-size: 12px;">
          导入时同名节点/客户端会被跳过，不会覆盖现有数据。
        </n-text>
      </n-space>
    </n-card>

    <!-- 数据库备份/恢复 -->
    <n-card style="margin-top: 16px;">
      <template #header>
        <span>数据库备份/恢复</span>
      </template>
      <n-space vertical>
        <n-text depth="3">备份完整数据库文件，包含所有配置和历史数据。</n-text>
        <n-space>
          <n-button :loading="backingUp" @click="handleBackup">
            下载备份
          </n-button>
          <n-upload
            :show-file-list="false"
            accept=".db"
            :custom-request="handleRestore"
          >
            <n-button :loading="restoring" type="warning">
              恢复备份
            </n-button>
          </n-upload>
        </n-space>
        <n-text depth="3" style="font-size: 12px; color: #e88;">
          注意：恢复备份后需要重启服务才能生效。
        </n-text>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage, useDialog } from 'naive-ui'
import { getSiteConfigs, updateSiteConfigs, exportData, importData, backupDatabase, restoreDatabase } from '../api'

const message = useMessage()
const dialog = useDialog()
const saving = ref(false)
const exporting = ref(false)
const importing = ref(false)
const backingUp = ref(false)
const restoring = ref(false)
const exportType = ref<'all' | 'nodes' | 'clients'>('all')

const exportTypeOptions = [
  { label: '全部', value: 'all' },
  { label: '仅节点', value: 'nodes' },
  { label: '仅客户端', value: 'clients' },
]

const roleOptions = [
  { label: '普通用户', value: 'user' },
  { label: '只读用户', value: 'viewer' },
]

const form = ref({
  site_name: '',
  site_description: '',
  site_url: '',
  favicon_url: '',
  logo_url: '',
  footer_text: '',
  custom_css: '',
  registration_enabled: false,
  email_verification_required: true,
  default_role: 'user',
  agent_auto_update: true,
  agent_force_update: false,
})

const loadConfigs = async () => {
  try {
    const data: any = await getSiteConfigs()
    form.value = {
      site_name: data.site_name || '',
      site_description: data.site_description || '',
      site_url: data.site_url || '',
      favicon_url: data.favicon_url || '',
      logo_url: data.logo_url || '',
      footer_text: data.footer_text || '',
      custom_css: data.custom_css || '',
      registration_enabled: data.registration_enabled === 'true',
      email_verification_required: data.email_verification_required !== 'false',
      default_role: data.default_role || 'user',
      agent_auto_update: data.agent_auto_update !== 'false',
      agent_force_update: data.agent_force_update === 'true',
    }
  } catch (e) {
    message.error('加载配置失败')
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    // 转换布尔值为字符串
    const saveData = {
      ...form.value,
      registration_enabled: form.value.registration_enabled ? 'true' : 'false',
      email_verification_required: form.value.email_verification_required ? 'true' : 'false',
      agent_auto_update: form.value.agent_auto_update ? 'true' : 'false',
      agent_force_update: form.value.agent_force_update ? 'true' : 'false',
    }
    await updateSiteConfigs(saveData)
    message.success('设置已保存，刷新页面生效')
    // 立即更新页面标题和图标
    document.title = form.value.site_name || 'GOST Panel'
    const favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement
    if (favicon && form.value.favicon_url) {
      favicon.href = form.value.favicon_url
    }
  } catch (e) {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

const onFaviconError = (e: Event) => {
  (e.target as HTMLImageElement).style.display = 'none'
}

const handleExport = async (format: 'json' | 'yaml') => {
  exporting.value = true
  try {
    const response: any = await exportData(format, exportType.value)
    // 创建下载链接
    const blob = new Blob([response], { type: format === 'yaml' ? 'application/x-yaml' : 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `gost-panel-export.${format}`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
    message.success('导出成功')
  } catch (e) {
    message.error('导出失败')
  } finally {
    exporting.value = false
  }
}

const handleImport = async ({ file }: { file: { file: File } }) => {
  importing.value = true
  try {
    const result: any = await importData(file.file)
    const msg = `导入完成: 节点 ${result.nodes_created} 个创建, ${result.nodes_skipped} 个跳过; 客户端 ${result.clients_created} 个创建, ${result.clients_skipped} 个跳过`
    message.success(msg)
  } catch (e) {
    message.error('导入失败')
  } finally {
    importing.value = false
  }
}

const handleBackup = async () => {
  backingUp.value = true
  try {
    const response: any = await backupDatabase()
    const blob = new Blob([response], { type: 'application/octet-stream' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `gost-panel-backup-${new Date().toISOString().slice(0, 10)}.db`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
    message.success('备份下载成功')
  } catch (e) {
    message.error('备份失败')
  } finally {
    backingUp.value = false
  }
}

const handleRestore = async ({ file }: { file: { file: File } }) => {
  dialog.warning({
    title: '确认恢复',
    content: '恢复备份将覆盖当前所有数据，此操作不可撤销。确定要继续吗？',
    positiveText: '确定恢复',
    negativeText: '取消',
    onPositiveClick: async () => {
      restoring.value = true
      try {
        await restoreDatabase(file.file)
        message.success('恢复成功，请重启服务以生效')
      } catch (e) {
        message.error('恢复失败')
      } finally {
        restoring.value = false
      }
    }
  })
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped>
</style>
