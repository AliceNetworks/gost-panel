<template>
  <div class="forgot-container">
    <!-- Background Orbs -->
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>
    <div class="bg-orb orb-3"></div>

    <n-card class="forgot-card">
      <div class="forgot-header">
        <h1 class="forgot-title">忘记密码</h1>
        <p class="forgot-subtitle">输入您的邮箱，我们将发送密码重置链接</p>
      </div>

      <!-- Success state -->
      <div v-if="submitted" class="forgot-content">
        <n-result status="info" title="邮件已发送" description="如果该邮箱已注册，您将收到密码重置链接。请检查您的邮箱（包括垃圾邮件文件夹）。">
          <template #footer>
            <n-button type="primary" @click="router.push('/login')">返回登录</n-button>
          </template>
        </n-result>
      </div>

      <!-- Form -->
      <n-form v-else ref="formRef" :model="form">
        <n-form-item label="邮箱">
          <n-input v-model:value="form.email" placeholder="user@example.com" @keyup.enter="handleSubmit" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="handleSubmit" class="forgot-btn">
          发送重置链接
        </n-button>
        <div class="back-link">
          <router-link to="/login">返回登录</router-link>
        </div>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { forgotPassword } from '../api'

const router = useRouter()
const message = useMessage()

const loading = ref(false)
const submitted = ref(false)
const form = ref({ email: '' })

const handleSubmit = async () => {
  if (!form.value.email) {
    message.error('请输入邮箱')
    return
  }

  loading.value = true
  try {
    await forgotPassword(form.value.email)
    submitted.value = true
  } catch (e: any) {
    // 为安全起见，无论是否找到邮箱都显示成功
    submitted.value = true
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.forgot-container {
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
  background: #f59e0b;
  top: -200px;
  right: -100px;
  animation: float 8s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: #ef4444;
  bottom: -150px;
  left: -100px;
  animation: float 10s ease-in-out infinite reverse;
}

.orb-3 {
  width: 300px;
  height: 300px;
  background: #f97316;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: pulse 6s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-30px) rotate(5deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 0.2; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 0.4; transform: translate(-50%, -50%) scale(1.1); }
}

.forgot-card {
  width: 400px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05) !important;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  z-index: 10;
}

.forgot-header {
  text-align: center;
  margin-bottom: 32px;
}

.forgot-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #f59e0b 0%, #ef4444 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 8px 0;
}

.forgot-subtitle {
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  margin: 0;
}

.forgot-content {
  padding: 20px 0;
}

.forgot-btn {
  margin-top: 8px;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 12px !important;
  background: linear-gradient(135deg, #f59e0b 0%, #ef4444 100%) !important;
  border: none !important;
  box-shadow: 0 4px 15px rgba(245, 158, 11, 0.4);
  transition: all 0.3s ease;
}

.forgot-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(245, 158, 11, 0.5);
}

.back-link {
  text-align: center;
  margin-top: 16px;
}

.back-link a {
  color: #3b82f6;
  text-decoration: none;
  font-size: 14px;
}

.back-link a:hover {
  text-decoration: underline;
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
