<template>
  <n-layout has-sider class="layout">
    <!-- 移动端遮罩层 -->
    <div
      v-if="isMobile && !collapsed"
      class="mobile-overlay"
      @click="collapsed = true"
    />

    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="isMobile ? 0 : 64"
      :width="200"
      :show-trigger="!isMobile"
      :collapsed="collapsed"
      :class="{ 'mobile-sider': isMobile }"
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <div class="logo">
        <span v-if="!collapsed">{{ siteConfig.site_name || 'GOST Panel' }}</span>
        <span v-else>{{ (siteConfig.site_name || 'G')[0] }}</span>
      </div>
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :value="currentMenu"
        @update:value="handleMenuSelect"
      />
      <div class="sidebar-version" :class="{ 'collapsed': collapsed }">
        <span v-if="!collapsed">{{ panelVersion }}</span>
        <span v-else>{{ panelVersion.split(' ').pop() }}</span>
      </div>
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered class="header">
        <div class="header-left">
          <!-- 移动端汉堡菜单 -->
          <n-button v-if="isMobile" quaternary circle class="menu-toggle" @click="collapsed = !collapsed">
            <template #icon>
              <n-icon size="22">
                <menu-outline />
              </n-icon>
            </template>
          </n-button>
          <div class="header-title">{{ currentTitle }}</div>
        </div>
        <div class="header-actions">
          <GlobalSearch v-if="!isMobile" />
          <n-dropdown :options="localeMenuOptions" @select="handleLocaleChange">
            <n-button quaternary circle>
              <template #icon>
                <n-icon><globe-outline /></n-icon>
              </template>
            </n-button>
          </n-dropdown>
          <n-button quaternary circle @click="themeStore.toggle">
            <template #icon>
              <n-icon>
                <moon-outline v-if="!themeStore.isDark" />
                <sunny-outline v-else />
              </n-icon>
            </template>
          </n-button>
          <n-dropdown :options="userOptions" @select="handleUserAction">
            <n-button quaternary class="user-btn">
              <span class="username">{{ userStore.user?.username || 'admin' }}</span>
            </n-button>
          </n-dropdown>
        </div>
      </n-layout-header>
      <n-layout-content class="content">
        <router-view />
      </n-layout-content>
    </n-layout>

    <!-- Change Password Modal -->
    <n-modal v-model:show="showPasswordModal" preset="dialog" :title="t('auth.changePassword')">
      <n-form :model="passwordForm" label-placement="left" label-width="100">
        <n-form-item :label="t('auth.oldPassword')">
          <n-input v-model:value="passwordForm.old_password" type="password" :placeholder="t('auth.oldPassword')" />
        </n-form-item>
        <n-form-item :label="t('auth.newPassword')">
          <n-input v-model:value="passwordForm.new_password" type="password" :placeholder="t('auth.newPassword')" />
        </n-form-item>
        <n-form-item :label="t('auth.confirmPassword')">
          <n-input v-model:value="passwordForm.confirm_password" type="password" :placeholder="t('auth.confirmPassword')" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showPasswordModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="changingPassword" @click="handleChangePassword">{{ t('common.confirm') }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Account Settings Modal -->
    <n-modal v-model:show="showAccountModal" preset="dialog" :title="t('auth.accountSettings')" style="width: 600px;">
      <n-tabs type="line" animated>
        <n-tab-pane name="profile" tab="个人信息">
          <n-form :model="profileForm" label-placement="left" label-width="100">
            <n-form-item :label="t('auth.username')">
              <n-input :value="userStore.user?.username" disabled />
            </n-form-item>
            <n-form-item :label="t('auth.email')">
              <n-input v-model:value="profileForm.email" placeholder="user@example.com" />
            </n-form-item>
          </n-form>
          <div style="margin-top: 16px; text-align: right;">
            <n-space>
              <n-button @click="showAccountModal = false">{{ t('common.cancel') }}</n-button>
              <n-button type="primary" :loading="savingProfile" @click="handleSaveProfile">{{ t('common.save') }}</n-button>
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="2fa" tab="双因素认证">
          <div v-if="!twoFactorEnabled">
            <n-alert type="info" title="未启用 2FA" style="margin-bottom: 16px;">
              启用双因素认证可以提高账户安全性。
            </n-alert>
            <n-button type="primary" @click="start2FASetup" :loading="loading2FA">启用 2FA</n-button>
          </div>

          <div v-else>
            <n-alert type="success" title="已启用 2FA" style="margin-bottom: 16px;">
              您的账户已受到双因素认证保护。
            </n-alert>
            <n-button type="warning" @click="show2FADisableModal = true">禁用 2FA</n-button>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-modal>

    <!-- 2FA Setup Modal -->
    <n-modal v-model:show="show2FASetupModal" preset="dialog" title="启用双因素认证" style="width: 500px;">
      <div v-if="!twoFactorVerified">
        <n-space vertical>
          <div style="text-align: center;">
            <p>1. 使用验证器 App 扫描二维码</p>
            <p style="font-size: 12px; color: #999;">推荐: Google Authenticator, Microsoft Authenticator</p>
            <img v-if="qrCode" :src="qrCode" style="max-width: 200px; margin: 16px 0;" />
          </div>
          <div style="text-align: center;">
            <p>或手动输入密钥:</p>
            <n-input :value="twoFactorSecret" readonly type="textarea" :autosize="{ minRows: 2, maxRows: 3 }" />
          </div>
          <div>
            <p>2. 输入验证器生成的 6 位数字验证码:</p>
            <n-input v-model:value="verifyCode" placeholder="000000" maxlength="6" />
          </div>
        </n-space>
      </div>
      <div v-else>
        <n-alert type="success" title="2FA 已启用" style="margin-bottom: 16px;">
          请妥善保存以下备份码，每个备份码只能使用一次。
        </n-alert>
        <n-space vertical>
          <div v-for="(code, idx) in backupCodes" :key="idx" style="font-family: monospace; padding: 4px 8px; background: rgba(128,128,128,0.1); border-radius: 4px;">
            {{ code }}
          </div>
        </n-space>
      </div>
      <template #action>
        <n-space v-if="!twoFactorVerified">
          <n-button @click="cancel2FASetup">取消</n-button>
          <n-button type="primary" :loading="loading2FA" @click="verify2FACode">验证并启用</n-button>
        </n-space>
        <n-space v-else>
          <n-button type="primary" @click="finish2FASetup">完成</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 2FA Disable Modal -->
    <n-modal v-model:show="show2FADisableModal" preset="dialog" title="禁用双因素认证">
      <n-alert type="warning" title="警告" style="margin-bottom: 16px;">
        禁用 2FA 将降低您的账户安全性。
      </n-alert>
      <n-form>
        <n-form-item label="当前密码">
          <n-input v-model:value="disable2FAPassword" type="password" placeholder="请输入当前密码以确认" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="show2FADisableModal = false">取消</n-button>
          <n-button type="warning" :loading="loading2FA" @click="confirmDisable2FA">确认禁用</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NIcon } from 'naive-ui'
import {
  HomeOutline,
  ServerOutline,
  DesktopOutline,
  PeopleOutline,
  LogOutOutline,
  KeyOutline,
  NotificationsOutline,
  SwapHorizontalOutline,
  GitNetworkOutline,
  SunnyOutline,
  MoonOutline,
  LinkOutline,
  SettingsOutline,
  ListOutline,
  MenuOutline,
  CardOutline,
  ShieldCheckmarkOutline,
  GlobeOutline,
} from '@vicons/ionicons5'
import { useUserStore } from '../stores/user'
import { useThemeStore } from '../stores/theme'
import { changePassword, getPublicSiteConfig, getProfile, updateProfile, getHealthInfo, enable2FA, verify2FA, disable2FA } from '../api'
import GlobalSearch from '../components/GlobalSearch.vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()
const message = useMessage()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const themeStore = useThemeStore()

// 移动端检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768

const checkMobile = () => {
  isMobile.value = window.innerWidth < MOBILE_BREAKPOINT
  // 移动端默认折叠侧边栏
  if (isMobile.value) {
    collapsed.value = true
  }
}

const collapsed = ref(false)
const panelVersion = ref('')
const showPasswordModal = ref(false)
const showAccountModal = ref(false)
const changingPassword = ref(false)
const savingProfile = ref(false)
const siteConfig = ref({
  site_name: 'GOST Panel',
  favicon_url: '/vite.svg',
  logo_url: '',
  footer_text: '',
})
const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: '',
})
const profileForm = ref({
  email: '',
})

// 2FA state
const show2FASetupModal = ref(false)
const show2FADisableModal = ref(false)
const loading2FA = ref(false)
const twoFactorEnabled = ref(false)
const twoFactorSecret = ref('')
const qrCode = ref('')
const verifyCode = ref('')
const twoFactorVerified = ref(false)
const backupCodes = ref<string[]>([])
const disable2FAPassword = ref('')

const renderIcon = (icon: any) => () => h(NIcon, null, { default: () => h(icon) })

const localeMenuOptions = computed(() => [
  { label: '中文', key: 'zh-CN' },
  { label: 'English', key: 'en' },
])

const handleLocaleChange = (key: string) => {
  locale.value = key
  localStorage.setItem('locale', key)
}

const menuOptions = computed(() => {
  const baseItems = [
    {
      label: t('menu.dashboard'),
      key: 'dashboard',
      icon: renderIcon(HomeOutline),
    },
    {
      label: t('menu.clients'),
      key: 'clients',
      icon: renderIcon(DesktopOutline),
    },
    {
      label: t('menu.nodes'),
      key: 'nodes',
      icon: renderIcon(ServerOutline),
    },
    {
      label: t('menu.portForwards'),
      key: 'port-forwards',
      icon: renderIcon(SwapHorizontalOutline),
    },
    {
      label: t('menu.nodeGroups'),
      key: 'node-groups',
      icon: renderIcon(GitNetworkOutline),
    },
    {
      label: t('menu.tunnels'),
      key: 'tunnels',
      icon: renderIcon(LinkOutline),
    },
  ]

  if (userStore.user?.role === 'admin') {
    baseItems.push(
      {
        label: t('menu.rules'),
        key: 'rules',
        icon: renderIcon(ShieldCheckmarkOutline),
      },
      {
        label: t('menu.users'),
        key: 'users',
        icon: renderIcon(PeopleOutline),
      },
      {
        label: t('menu.notify'),
        key: 'notify',
        icon: renderIcon(NotificationsOutline),
      },
      {
        label: t('menu.operationLogs'),
        key: 'operation-logs',
        icon: renderIcon(ListOutline),
      },
      {
        label: t('menu.plans'),
        key: 'plans',
        icon: renderIcon(CardOutline),
      },
      {
        label: t('menu.settings'),
        key: 'settings',
        icon: renderIcon(SettingsOutline),
      }
    )
  }

  return baseItems
})

const userOptions = computed(() => [
  {
    label: t('auth.accountSettings'),
    key: 'account-settings',
    icon: renderIcon(SettingsOutline),
  },
  {
    label: t('auth.changePassword'),
    key: 'change-password',
    icon: renderIcon(KeyOutline),
  },
  {
    label: t('auth.logout'),
    key: 'logout',
    icon: renderIcon(LogOutOutline),
  },
])

const currentMenu = computed(() => route.name as string)

const currentTitle = computed(() => {
  const menu = menuOptions.value.find((m) => m.key === currentMenu.value)
  return menu?.label || t('menu.dashboard')
})

const handleMenuSelect = (key: string) => {
  console.log('[Menu] Selected:', key, '| Current:', currentMenu.value)
  if (key === currentMenu.value) {
    console.log('[Menu] Same as current, skipping')
    // 移动端点击同一菜单也折叠侧边栏
    if (isMobile.value) {
      collapsed.value = true
    }
    return
  }
  console.log('[Menu] Navigating to:', key)
  router.push({ name: key }).then(() => {
    console.log('[Menu] Navigation success')
    // 移动端导航后自动折叠侧边栏
    if (isMobile.value) {
      collapsed.value = true
    }
  }).catch((err) => {
    console.error('[Menu] Navigation error:', err)
  })
}

const handleUserAction = async (key: string) => {
  if (key === 'logout') {
    userStore.logout()
    router.push('/login')
  } else if (key === 'change-password') {
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
    showPasswordModal.value = true
  } else if (key === 'account-settings') {
    await loadProfile()
    showAccountModal.value = true
  }
}

const loadProfile = async () => {
  try {
    const user: any = await getProfile()
    profileForm.value = {
      email: user.email || '',
    }
    twoFactorEnabled.value = user.two_factor_enabled || false
  } catch {
    message.error(t('auth.loadProfileFailed'))
  }
}

const handleSaveProfile = async () => {
  savingProfile.value = true
  try {
    await updateProfile({ email: profileForm.value.email })
    message.success(t('auth.accountUpdated'))
    showAccountModal.value = false
  } catch (e: any) {
    message.error(e.response?.data?.error || t('auth.saveFailed'))
  } finally {
    savingProfile.value = false
  }
}

const handleChangePassword = async () => {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    message.error(t('auth.passwordMismatch'))
    return
  }

  changingPassword.value = true
  try {
    await changePassword(passwordForm.value.old_password, passwordForm.value.new_password)
    message.success(t('auth.passwordChanged'))
    showPasswordModal.value = false
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (e: any) {
    message.error(e.response?.data?.error || t('auth.loginFailed'))
  } finally {
    changingPassword.value = false
  }
}

// 2FA functions
const start2FASetup = async () => {
  loading2FA.value = true
  try {
    const res: any = await enable2FA()
    twoFactorSecret.value = res.secret
    qrCode.value = res.qrcode
    show2FASetupModal.value = true
  } catch (e: any) {
    message.error(e.response?.data?.error || '启用 2FA 失败')
  } finally {
    loading2FA.value = false
  }
}

const verify2FACode = async () => {
  if (!verifyCode.value || verifyCode.value.length !== 6) {
    message.error('请输入 6 位验证码')
    return
  }

  loading2FA.value = true
  try {
    const res: any = await verify2FA(verifyCode.value)
    backupCodes.value = res.backup_codes || []
    twoFactorVerified.value = true
    twoFactorEnabled.value = true
    message.success('2FA 已启用')
  } catch (e: any) {
    message.error(e.response?.data?.error || '验证码错误')
  } finally {
    loading2FA.value = false
  }
}

const finish2FASetup = () => {
  show2FASetupModal.value = false
  twoFactorVerified.value = false
  twoFactorSecret.value = ''
  qrCode.value = ''
  verifyCode.value = ''
  backupCodes.value = []
  loadProfile()
}

const cancel2FASetup = () => {
  show2FASetupModal.value = false
  twoFactorVerified.value = false
  twoFactorSecret.value = ''
  qrCode.value = ''
  verifyCode.value = ''
}

const confirmDisable2FA = async () => {
  if (!disable2FAPassword.value) {
    message.error('请输入密码')
    return
  }

  loading2FA.value = true
  try {
    await disable2FA(disable2FAPassword.value)
    twoFactorEnabled.value = false
    show2FADisableModal.value = false
    disable2FAPassword.value = ''
    message.success('2FA 已禁用')
  } catch (e: any) {
    message.error(e.response?.data?.error || '禁用失败')
  } finally {
    loading2FA.value = false
  }
}

// 加载网站配置
const loadSiteConfig = async () => {
  try {
    const config = await getPublicSiteConfig()
    siteConfig.value = config
    // 更新页面标题
    if (config.site_name) {
      document.title = config.site_name
    }
    // 更新 favicon
    if (config.favicon_url) {
      let favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement
      if (!favicon) {
        favicon = document.createElement('link')
        favicon.rel = 'icon'
        document.head.appendChild(favicon)
      }
      favicon.href = config.favicon_url
    }
    // 注入自定义 CSS
    if (config.custom_css) {
      let style = document.getElementById('custom-css') as HTMLStyleElement
      if (!style) {
        style = document.createElement('style')
        style.id = 'custom-css'
        document.head.appendChild(style)
      }
      style.textContent = config.custom_css
    }
  } catch {
    // Site config loading is non-critical, silently fail
  }
}

// 加载版本信息
const loadVersion = async () => {
  try {
    const data = await getHealthInfo()
    panelVersion.value = `GOST Panel ${data.version || ''}`
  } catch {
    panelVersion.value = 'GOST Panel'
  }
}

onMounted(() => {
  loadSiteConfig()
  loadVersion()
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.layout {
  height: 100vh;
  background: transparent;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: -0.5px;
}

.sidebar-version {
  padding: 12px 16px;
  font-size: 11px;
  color: rgba(128, 128, 128, 0.6);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: auto;
}

.sidebar-version.collapsed {
  padding: 12px 4px;
  font-size: 10px;
}

:deep(.n-layout-sider-scroll-container) {
  display: flex;
  flex-direction: column;
}

.header {
  height: 64px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
}

.content {
  padding: 24px;
  background: transparent;
  min-height: calc(100vh - 64px);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 移动端遮罩 */
.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

/* 移动端侧边栏 */
.mobile-sider {
  position: fixed !important;
  left: 0;
  top: 0;
  height: 100vh;
  z-index: 1000;
  transition: transform 0.3s ease;
}

.mobile-sider:deep(.n-layout-sider-scroll-container) {
  height: 100%;
}

/* 暗色模式样式 */
:global(html.dark) :deep(.n-layout-sider) {
  background: rgba(255, 255, 255, 0.03) !important;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

:global(html.dark) :deep(.n-layout-header) {
  background: rgba(255, 255, 255, 0.03) !important;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

:global(html.dark) .header-title {
  color: rgba(255, 255, 255, 0.9);
}

/* 亮色模式样式 - 柔和护眼 */
:global(html:not(.dark)) :deep(.n-layout-sider) {
  background: #ffffff !important;
  border-right: 1px solid #e8e4db;
}

:global(html:not(.dark)) :deep(.n-layout-header) {
  background: #ffffff !important;
  border-bottom: 1px solid #e8e4db;
}

:global(html:not(.dark)) .header-title {
  color: #2c3e50;
}

:global(html:not(.dark)) .content {
  background: #f8f6f1 !important;
}

:deep(.n-menu-item-content--selected) {
  background: rgba(59, 130, 246, 0.15) !important;
}

:deep(.n-menu-item-content--selected::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: linear-gradient(180deg, #3b82f6, #8b5cf6);
  border-radius: 0 2px 2px 0;
}

/* 移动端响应式 */
@media (max-width: 768px) {
  .header {
    padding: 0 12px;
  }

  .header-title {
    font-size: 16px;
  }

  .content {
    padding: 12px;
  }

  .username {
    max-width: 60px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .user-btn {
    padding: 0 8px !important;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 0 8px;
  }

  .header-title {
    font-size: 14px;
  }

  .content {
    padding: 8px;
  }
}
</style>
