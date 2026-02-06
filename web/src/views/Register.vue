<template>
  <div class="register-container">
    <!-- Background Orbs -->
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>
    <div class="bg-orb orb-3"></div>

    <n-card class="register-card">
      <div class="register-header">
        <img v-if="siteConfig.logo_url" :src="siteConfig.logo_url" class="register-logo" alt="Logo" />
        <h1 class="register-title">{{ siteConfig.site_name || 'GOST Panel' }}</h1>
        <p class="register-subtitle">创建新账户</p>
      </div>

      <!-- Registration disabled message -->
      <div v-if="!registrationEnabled" class="registration-disabled">
        <n-result status="warning" title="注册已关闭" description="当前不允许新用户注册，请联系管理员">
          <template #footer>
            <n-button @click="router.push('/login')">返回登录</n-button>
          </template>
        </n-result>
      </div>

      <!-- Registration success message -->
      <div v-else-if="registered" class="registration-success">
        <n-result status="success" :title="emailVerificationRequired ? '请验证邮箱' : '注册成功'" :description="successMessage">
          <template #footer>
            <n-button type="primary" @click="router.push('/login')">前往登录</n-button>
          </template>
        </n-result>
      </div>

      <!-- Registration form -->
      <n-form v-else ref="formRef" :model="form" :rules="rules">
        <n-form-item path="username" label="用户名">
          <n-input v-model:value="form.username" placeholder="请输入用户名 (至少3个字符)" />
        </n-form-item>
        <n-form-item path="email" label="邮箱">
          <n-input v-model:value="form.email" placeholder="user@example.com" />
        </n-form-item>
        <n-form-item path="password" label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="请输入密码 (至少6个字符)"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item path="confirmPassword" label="确认密码">
          <n-input
            v-model:value="form.confirmPassword"
            type="password"
            placeholder="再次输入密码"
            show-password-on="click"
            @keyup.enter="handleRegister"
          />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="handleRegister" class="register-btn">
          注册
        </n-button>
        <div class="login-link">
          已有账户？<router-link to="/login">立即登录</router-link>
        </div>
      </n-form>

      <div v-if="siteConfig.footer_text" class="register-footer">
        {{ siteConfig.footer_text }}
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { getPublicSiteConfig, register, getRegistrationStatus } from '../api'

const router = useRouter()
const message = useMessage()

const loading = ref(false)
const registered = ref(false)
const registrationEnabled = ref(true)
const emailVerificationRequired = ref(true)
const successMessage = ref('')

const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const siteConfig = ref({
  site_name: 'GOST Panel',
  site_description: '代理服务管理平台',
  logo_url: '',
  favicon_url: '',
  footer_text: '',
})

const validateConfirmPassword = (_rule: any, value: string) => {
  if (value !== form.value.password) {
    return new Error('两次输入的密码不一致')
  }
  return true
}

const rules = {
  username: [
    { required: true, message: '请输入用户名' },
    { min: 3, max: 50, message: '用户名长度应为3-50个字符' }
  ],
  email: [
    { required: true, message: '请输入邮箱' },
    { type: 'email', message: '请输入有效的邮箱地址' }
  ],
  password: [
    { required: true, message: '请输入密码' },
    { min: 6, message: '密码至少6个字符' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码' },
    { validator: validateConfirmPassword, trigger: ['blur', 'input'] }
  ],
}

const loadSiteConfig = async () => {
  try {
    const config = await getPublicSiteConfig()
    siteConfig.value = { ...siteConfig.value, ...config }
    if (config.site_name) {
      document.title = config.site_name + ' - 注册'
    }
  } catch {
    // Site config loading is non-critical
  }
}

const checkRegistrationStatus = async () => {
  try {
    const status: any = await getRegistrationStatus()
    registrationEnabled.value = status.enabled
    emailVerificationRequired.value = status.email_verification
  } catch {
    // Registration status check is non-critical
  }
}

const handleRegister = async () => {
  if (form.value.password !== form.value.confirmPassword) {
    message.error('两次输入的密码不一致')
    return
  }

  loading.value = true
  try {
    const result: any = await register(form.value.username, form.value.email, form.value.password)
    registered.value = true

    if (result.email_verification) {
      successMessage.value = '验证邮件已发送到您的邮箱，请查收并点击验证链接完成注册。'
    } else {
      successMessage.value = '您的账户已创建，现在可以登录了。'
    }
  } catch (e: any) {
    message.error(e.response?.data?.error || '注册失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSiteConfig()
  checkRegistrationStatus()
})
</script>

<style scoped>
.register-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0a0e27;
  position: relative;
  overflow: hidden;
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
  background: #10b981;
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

.register-card {
  width: 420px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05) !important;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  z-index: 10;
}

.register-header {
  text-align: center;
  margin-bottom: 32px;
}

.register-logo {
  max-width: 80px;
  max-height: 80px;
  margin-bottom: 16px;
  border-radius: 12px;
}

.register-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #10b981 0%, #3b82f6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 8px 0;
}

.register-subtitle {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin: 0;
}

.register-footer {
  text-align: center;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
}

.register-btn {
  margin-top: 8px;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 12px !important;
  background: linear-gradient(135deg, #10b981 0%, #3b82f6 100%) !important;
  border: none !important;
  box-shadow: 0 4px 15px rgba(16, 185, 129, 0.4);
  transition: all 0.3s ease;
}

.register-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.5);
}

.login-link {
  text-align: center;
  margin-top: 16px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
}

.login-link a {
  color: #3b82f6;
  text-decoration: none;
}

.login-link a:hover {
  text-decoration: underline;
}

.registration-disabled,
.registration-success {
  padding: 20px 0;
}

:deep(.n-form-item-label) {
  color: rgba(255, 255, 255, 0.7) !important;
}

:deep(.n-result) {
  background: transparent;
}

:deep(.n-result-header__title) {
  color: rgba(255, 255, 255, 0.9) !important;
}

:deep(.n-result-header__description) {
  color: rgba(255, 255, 255, 0.6) !important;
}
</style>
