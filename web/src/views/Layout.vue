<template>
  <n-layout has-sider class="layout">
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="200"
      show-trigger
      :collapsed="collapsed"
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
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered class="header">
        <div class="header-title">{{ currentTitle }}</div>
        <div class="header-actions">
          <GlobalSearch />
          <n-button quaternary circle @click="themeStore.toggle">
            <template #icon>
              <n-icon>
                <moon-outline v-if="!themeStore.isDark" />
                <sunny-outline v-else />
              </n-icon>
            </template>
          </n-button>
          <n-dropdown :options="userOptions" @select="handleUserAction">
            <n-button quaternary>
              {{ userStore.user?.username || 'admin' }}
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
import { ref, computed, h, onMounted } from 'vue'
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
} from '@vicons/ionicons5'
import { useUserStore } from '../stores/user'
import { useThemeStore } from '../stores/theme'
import { changePassword, getPublicSiteConfig, getProfile, updateProfile } from '../api'
import GlobalSearch from '../components/GlobalSearch.vue'
import { useMessage } from 'naive-ui'

const message = useMessage()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const themeStore = useThemeStore()

const collapsed = ref(false)
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
  router.push({ name: key })
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

onMounted(() => {
  loadSiteConfig()
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

.header {
  height: 64px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
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
</style>
