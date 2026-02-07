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
    <n-modal v-model:show="showPasswordModal" preset="dialog" title="修改密码">
      <n-form :model="passwordForm" label-placement="left" label-width="100">
        <n-form-item label="当前密码">
          <n-input v-model:value="passwordForm.old_password" type="password" placeholder="请输入当前密码" />
        </n-form-item>
        <n-form-item label="新密码">
          <n-input v-model:value="passwordForm.new_password" type="password" placeholder="请输入新密码" />
        </n-form-item>
        <n-form-item label="确认密码">
          <n-input v-model:value="passwordForm.confirm_password" type="password" placeholder="请再次输入新密码" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showPasswordModal = false">取消</n-button>
          <n-button type="primary" :loading="changingPassword" @click="handleChangePassword">确定</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Account Settings Modal -->
    <n-modal v-model:show="showAccountModal" preset="dialog" title="账户设置" style="width: 500px;">
      <n-form :model="profileForm" label-placement="left" label-width="100">
        <n-form-item label="用户名">
          <n-input :value="userStore.user?.username" disabled />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="profileForm.email" placeholder="user@example.com" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAccountModal = false">取消</n-button>
          <n-button type="primary" :loading="savingProfile" @click="handleSaveProfile">保存</n-button>
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
} from '@vicons/ionicons5'
import { useUserStore } from '../stores/user'
import { useThemeStore } from '../stores/theme'
import { changePassword, getPublicSiteConfig, getProfile, updateProfile, getHealthInfo } from '../api'
import GlobalSearch from '../components/GlobalSearch.vue'
import { useMessage } from 'naive-ui'

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

const renderIcon = (icon: any) => () => h(NIcon, null, { default: () => h(icon) })

const menuOptions = [
  {
    label: '仪表盘',
    key: 'dashboard',
    icon: renderIcon(HomeOutline),
  },
  {
    label: '客户端',
    key: 'clients',
    icon: renderIcon(DesktopOutline),
  },
  {
    label: '节点管理',
    key: 'nodes',
    icon: renderIcon(ServerOutline),
  },
  {
    label: '端口转发',
    key: 'port-forwards',
    icon: renderIcon(SwapHorizontalOutline),
  },
  {
    label: '负载均衡',
    key: 'node-groups',
    icon: renderIcon(GitNetworkOutline),
  },
  {
    label: '隧道转发',
    key: 'tunnels',
    icon: renderIcon(LinkOutline),
  },
  {
    label: '规则管理',
    key: 'rules',
    icon: renderIcon(ShieldCheckmarkOutline),
  },
  {
    label: '用户管理',
    key: 'users',
    icon: renderIcon(PeopleOutline),
  },
  {
    label: '告警通知',
    key: 'notify',
    icon: renderIcon(NotificationsOutline),
  },
  {
    label: '操作日志',
    key: 'operation-logs',
    icon: renderIcon(ListOutline),
  },
  {
    label: '套餐管理',
    key: 'plans',
    icon: renderIcon(CardOutline),
  },
  {
    label: '网站设置',
    key: 'settings',
    icon: renderIcon(SettingsOutline),
  },
]

const userOptions = [
  {
    label: '账户设置',
    key: 'account-settings',
    icon: renderIcon(SettingsOutline),
  },
  {
    label: '修改密码',
    key: 'change-password',
    icon: renderIcon(KeyOutline),
  },
  {
    label: '退出登录',
    key: 'logout',
    icon: renderIcon(LogOutOutline),
  },
]

const currentMenu = computed(() => route.name as string)

const currentTitle = computed(() => {
  const menu = menuOptions.find((m) => m.key === currentMenu.value)
  return menu?.label || '仪表盘'
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
  } catch {
    message.error('加载用户信息失败')
  }
}

const handleSaveProfile = async () => {
  savingProfile.value = true
  try {
    await updateProfile({ email: profileForm.value.email })
    message.success('账户信息已更新')
    showAccountModal.value = false
  } catch (e: any) {
    message.error(e.response?.data?.error || '保存失败')
  } finally {
    savingProfile.value = false
  }
}

const handleChangePassword = async () => {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    message.error('两次输入的密码不一致')
    return
  }

  changingPassword.value = true
  try {
    await changePassword(passwordForm.value.old_password, passwordForm.value.new_password)
    message.success('密码修改成功')
    showPasswordModal.value = false
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (e: any) {
    message.error(e.response?.data?.error || '密码修改失败')
  } finally {
    changingPassword.value = false
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
