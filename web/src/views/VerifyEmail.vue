<template>
  <div class="verify-container">
    <!-- Background Orbs -->
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>
    <div class="bg-orb orb-3"></div>

    <n-card class="verify-card">
      <div class="verify-header">
        <h1 class="verify-title">邮箱验证</h1>
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="verify-content">
        <n-spin size="large" />
        <p class="verify-message">正在验证您的邮箱...</p>
      </div>

      <!-- Success state -->
      <div v-else-if="verified" class="verify-content">
        <n-result status="success" title="验证成功" description="您的邮箱已验证，现在可以登录了">
          <template #footer>
            <n-button type="primary" @click="router.push('/login')">前往登录</n-button>
          </template>
        </n-result>
      </div>

      <!-- Error state -->
      <div v-else class="verify-content">
        <n-result status="error" title="验证失败" :description="errorMessage">
          <template #footer>
            <n-space>
              <n-button @click="router.push('/login')">返回登录</n-button>
              <n-button type="primary" @click="router.push('/register')">重新注册</n-button>
            </n-space>
          </template>
        </n-result>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { verifyEmail } from '../api'

const router = useRouter()
const route = useRoute()

const loading = ref(true)
const verified = ref(false)
const errorMessage = ref('验证链接无效或已过期')

const doVerify = async () => {
  const token = route.query.token as string
  if (!token) {
    loading.value = false
    errorMessage.value = '验证令牌缺失'
    return
  }

  try {
    await verifyEmail(token)
    verified.value = true
  } catch (e: any) {
    errorMessage.value = e.response?.data?.error || '验证链接无效或已过期'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  doVerify()
})
</script>

<style scoped>
.verify-container {
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
  background: #10b981;
  top: -200px;
  right: -100px;
  animation: float 8s ease-in-out infinite;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: #3b82f6;
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
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-30px) rotate(5deg); }
}

@keyframes pulse {
  0%, 100% { opacity: 0.2; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 0.4; transform: translate(-50%, -50%) scale(1.1); }
}

.verify-card {
  width: 420px;
  padding: 40px 20px;
  background: rgba(255, 255, 255, 0.05) !important;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  z-index: 10;
}

.verify-header {
  text-align: center;
  margin-bottom: 32px;
}

.verify-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #10b981 0%, #3b82f6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.verify-content {
  text-align: center;
}

.verify-message {
  color: rgba(255, 255, 255, 0.6);
  margin-top: 16px;
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
