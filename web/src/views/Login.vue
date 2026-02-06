<template>
  <div class="login-container">
    <!-- Background Orbs -->
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>
    <div class="bg-orb orb-3"></div>

    <n-card class="login-card">
      <div class="login-header">
        <img v-if="siteConfig.logo_url" :src="siteConfig.logo_url" class="login-logo" alt="Logo" />
        <h1 class="login-title">{{ siteConfig.site_name || 'GOST Panel' }}</h1>
        <p class="login-subtitle">{{ siteConfig.site_description || '代理服务管理平台' }}</p>
      </div>
      <n-form ref="formRef" :model="form" :rules="rules">
        <n-form-item path="username" label="用户名">
          <n-input v-model:value="form.username" placeholder="admin" />
        </n-form-item>
        <n-form-item path="password" label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="请输入密码"
            @keyup.enter="handleLogin"
          />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="handleLogin" class="login-btn">
          登录
        </n-button>
      </n-form>
      <div class="login-links">
        <router-link to="/forgot-password">忘记密码？</router-link>
        <span v-if="registrationEnabled">|</span>
        <router-link v-if="registrationEnabled" to="/register">注册账户</router-link>
      </div>
      <div v-if="siteConfig.footer_text" class="login-footer">
        {{ siteConfig.footer_text }}
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useUserStore } from '../stores/user'
import { getPublicSiteConfig, getRegistrationStatus } from '../api'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const loading = ref(false)
const registrationEnabled = ref(false)
const form = ref({
  username: '',
  password: '',
})

const siteConfig = ref({
  site_name: 'GOST Panel',
  site_description: '代理服务管理平台',
  logo_url: '',
  favicon_url: '',
  footer_text: '',
})

const rules = {
  username: { required: true, message: '请输入用户名' },
  password: { required: true, message: '请输入密码' },
}

const loadSiteConfig = async () => {
  try {
    const config = await getPublicSiteConfig()
    siteConfig.value = { ...siteConfig.value, ...config }
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
  } catch {
    // Site config loading is non-critical
  }
}

const handleLogin = async () => {
  loading.value = true
  try {
    await userStore.login(form.value.username, form.value.password)
    message.success('登录成功')
    // 检查是否需要强制修改密码
    if (userStore.user && !userStore.user.password_changed) {
      message.warning('首次登录请修改默认密码')
      router.push('/change-password?force=1')
    } else {
      router.push('/')
    }
  } catch (e: any) {
    message.error(e.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSiteConfig()
  checkRegistrationStatus()
})

const checkRegistrationStatus = async () => {
  try {
    const status: any = await getRegistrationStatus()
    registrationEnabled.value = status.enabled
  } catch {
    // Registration status check is non-critical
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0a0e27;
  position: relative;
  overflow: hidden;
}

.login-links {
  text-align: center;
  margin-top: 16px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
}

.login-links a {
  color: #3b82f6;
  text-decoration: none;
  margin: 0 8px;
}

.login-links a:hover {
  text-decoration: underline;
}

.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.4;
  pointer-events: none;
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: #3b82f6;
  top: -200px;
  right: -100px;
  animation: float 8s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: #8b5cf6;
  bottom: -150px;
  left: -100px;
  animation: float 10s ease-in-out infinite reverse;
}

.orb-3 {
  width: 300px;
  height: 300px;
  background: #06b6d4;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: pulse 6s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-30px) rotate(5deg);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.2;
    transform: translate(-50%, -50%) scale(1);
  }
  50% {
    opacity: 0.4;
    transform: translate(-50%, -50%) scale(1.1);
  }
}

.login-card {
  width: 400px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05) !important;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  z-index: 10;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo {
  max-width: 80px;
  max-height: 80px;
  margin-bottom: 16px;
  border-radius: 12px;
}

.login-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 8px 0;
}

.login-subtitle {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin: 0;
}

.login-footer {
  text-align: center;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
}

.login-btn {
  margin-top: 8px;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 12px !important;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%) !important;
  border: none !important;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.4);
  transition: all 0.3s ease;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.5);
}

:deep(.n-form-item-label) {
  color: rgba(255, 255, 255, 0.7) !important;
}
</style>
